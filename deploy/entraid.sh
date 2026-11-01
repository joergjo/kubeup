#!/bin/bash
set -eE -o pipefail
trap 'echo "Entra setup failed at line $LINENO. Check deploy/entraid.log and the Azure CLI error above." >&2; exit 1' ERR

config_file=../.env.entraid
graph_url=https://graph.microsoft.com/v1.0

if [[ -f "$config_file" ]]; then
    echo "The .env.entraid file already exists. Using existing Entra ID application." > ./entraid.log
    cat "$config_file"
    exit 0
fi

for tool in az jq uuidgen; do
    command -v "$tool" >/dev/null || {
        echo "Required command not found: $tool" >&2
        exit 1
    }
done

graph() {
    az rest --method "$1" --url "$2" --output json "${@:3}"
}

required_id() {
    jq -esr 'select(length == 1) | .[0].id | strings |
        select(test("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"))'
}

find_by_app_id() {
    local filter response
    filter=$(jq -nr --arg id "$2" '"appId eq \u0027" + $id + "\u0027" | @uri')
    response=$(graph GET "$graph_url/$1?\$filter=$filter")
    echo "$response" | jq -e '.value | arrays | select(length <= 1)' >/dev/null
    echo "$response" | jq -c '.value[0] // empty'
}

ensure_principal() {
    local principal body
    principal=$(find_by_app_id servicePrincipals "$1")
    if [[ -z "$principal" ]]; then
        body=$(jq -n --arg id "$1" '{appId: $id}')
        principal=$(graph POST "$graph_url/servicePrincipals" --body "$body")
    fi
    echo "$principal" | required_id
}

ensure_assignment() {
    local principal_id="$1" url response next body
    url="$graph_url/servicePrincipals/$webhook_sp_id/appRoleAssignedTo"
    while [[ -n "$url" ]]; do
        response=$(graph GET "$url")
        echo "$response" | jq -e '.value | arrays' >/dev/null
        if echo "$response" | jq -e --arg principal "$principal_id" --arg role "$app_role_id" \
            --arg resource "$webhook_sp_id" \
            '.value[] | select((.principalId | ascii_downcase) == ($principal | ascii_downcase) and
                (.appRoleId | ascii_downcase) == ($role | ascii_downcase) and
                (.resourceId | ascii_downcase) == ($resource | ascii_downcase))' >/dev/null; then
            return
        fi
        next=$(echo "$response" | jq -r '."@odata.nextLink" // empty')
        if [[ -n "$next" && "$next" != "$graph_url/"* ]]; then
            echo "Unexpected Graph pagination URL." >&2
            exit 1
        fi
        url="$next"
    done
    body=$(jq -n --arg principal "$principal_id" --arg resource "$webhook_sp_id" \
        --arg role "$app_role_id" '{principalId: $principal, resourceId: $resource, appRoleId: $role}')
    graph POST "$graph_url/servicePrincipals/$webhook_sp_id/appRoleAssignedTo" --body "$body" |
        jq -er '.id | strings | select(length > 0)' >/dev/null
}

{
    account=$(az account show --output json)
    if [[ $(echo "$account" | jq -r '.user.type') != user ]]; then
        echo "Interactive Azure CLI login required. Run az login --tenant <tenant-id>." >&2
        exit 1
    fi
    current_user=$(graph GET "$graph_url/me")
    current_user_id=$(echo "$current_user" | required_id)

    webhook_app_id=${1:-}
    if [[ -n "$webhook_app_id" ]]; then
        if ! [[ "$webhook_app_id" =~ ^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$ ]]; then
            echo "Webhook application ID must be a UUID." >&2
            exit 1
        fi
        webhook_app=$(find_by_app_id applications "$webhook_app_id")
    else
        body=$(jq -n --arg name "kubeup-webhook-$(date +%s)" '{displayName: $name}')
        webhook_app=$(graph POST "$graph_url/applications" --body "$body")
        webhook_app_id=$(echo "$webhook_app" | jq '{id: .appId}' | required_id)
    fi
    webhook_object_id=$(echo "$webhook_app" | required_id)

    # The Event Grid application ID is specific to the public Azure cloud.
    eg_app_id=4962773b-9cdb-44cf-a8bf-237846a00ab7
    eg_sp_id=$(ensure_principal "$eg_app_id")
    webhook_sp_id=$(ensure_principal "$webhook_app_id")

    role_name=AzureEventGridSecureWebhookSubscriber
    app_roles=$(echo "$webhook_app" | jq -ce '.appRoles | arrays')
    app_role_id=$(echo "$app_roles" | jq -r --arg name "$role_name" '.[] | select(.value == $name) | .id')
    if [[ -z "$app_role_id" ]]; then
        app_role_id=$(uuidgen | tr '[:upper:]' '[:lower:]')
        body=$(echo "$app_roles" | jq --arg name "$role_name" --arg id "$app_role_id" \
            '{appRoles: (. + [{allowedMemberTypes: ["User", "Application"], description: "Azure Event Grid Role",
                displayName: $name, id: $id, isEnabled: true, value: $name}])}')
        graph PATCH "$graph_url/applications/$webhook_object_id" --body "$body" >/dev/null
    fi
    echo "$app_roles" | jq -e --arg name "$role_name" \
        '[.[] | select(.value == $name)] | length <= 1' >/dev/null
    app_role_id=$(jq -n --arg id "$app_role_id" '{id: $id}' | required_id | tr '[:upper:]' '[:lower:]')

    # Wait for the application role to propagate to its service principal.
    for attempt in 1 2 3 4 5; do
        principal=$(graph GET "$graph_url/servicePrincipals/$webhook_sp_id")
        if echo "$principal" | jq -e --arg id "$app_role_id" \
            '.appRoles[] | select((.id | ascii_downcase) == $id and .isEnabled == true and
                (.allowedMemberTypes | index("User")) != null and
                (.allowedMemberTypes | index("Application")) != null)' >/dev/null; then
            break
        fi
        if [[ "$attempt" == 5 ]]; then
            echo "Required app role is not available on the webhook service principal. Retry setup with application ID $webhook_app_id." >&2
            exit 1
        fi
        sleep 2
    done

    ensure_assignment "$current_user_id"
    ensure_assignment "$eg_sp_id"
    echo "Entra setup completed successfully."
} > ./entraid.log

printf "KU_APP_ID='%s'\n" "$webhook_app_id" > "$config_file"
cat "$config_file"
