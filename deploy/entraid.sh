#!/bin/bash

if [[ -f ./.env.entraid ]]; then
    echo "The .env.entraid file already exists. Using existing Entra ID application." > ./entraid.log
    cat ./.env.entraid
    exit 0
fi

{
    webhook_app_id=$1

    if [[ -z "$webhook_app_id" ]]; then
        echo "Webhook application ID has not been provided. Creating new webhook application and service principal..."
        webhook_app=$(mgc applications create --body "{\
            \"displayName\": \"kubeup-webhook-$(date +%s)\"\
        }\
        ")
        webhook_app_id=$(echo $webhook_app | jq -r ".appId")
        mgc service-principals create --output NONE --body "{\
            \"appId\": \"$webhook_app_id\"\
        }\
        "
    fi

    # Creates Azure Event Grid Microsoft Entra Application if not exists
    # You don't need to modify this id
    # But Azure Event Grid Microsoft Entra Application Id is different for different clouds
    eg_app_id="4962773b-9cdb-44cf-a8bf-237846a00ab7"
    eg_sp=$(mgc service-principals-with-app-id get --app-id "$eg_app_id" --select "id, appDisplayName")
    eg_sp_id=$(echo "$eg_sp" | jq -r ".id")
    eg_display_name=$(echo "$eg_sp" | jq -r ".appDisplayName")

    if [[ $eg_display_name=="Microsoft.EventGrid" ]]; then
        echo "Azure Event Grid Microsoft Entra Application already exists"
    else
        echo "Creating the Azure Event Grid Microsoft Entra Application"
        mgc service-principals create --body "{\
            \"appId\": \"$eg_app_id\"\
        }\
        "
    fi

    # Creates the Azure app role for the webhook Microsoft Entra application
    eg_role_name="AzureEventGridSecureWebhookSubscriber"
    if [[ -z $webhook_app ]]; then
        webhook_app=$(mgc applications-with-app-id get --app-id "$webhook_app_id")
    fi

    echo "Microsoft Entra App roles before addition of the new role..."
    app_roles=$(echo $webhook_app | jq -r '.appRoles')
    echo $app_roles | jq -r ".[].displayName"

    app_role=$(echo $app_roles | jq -r ".[] | select(.displayName == \"$eg_role_name\")")

    if [[ -z $app_role ]]; then
        echo "Adding the new role to the Microsoft Entra Application"
        app_role_id=$(uuidgen)
        updated_app_roles=$(echo $webhook_app | jq -r ".appRoles + [{\
            \"allowedMemberTypes\": [\
                \"User\",\
                \"Application\"\
            ],\
            \"description\": \"Azure Event Grid Role\",\
            \"displayName\": \"$eg_role_name\",\
            \"id\": \"$app_role_id\",\
            \"isEnabled\": true,\
            \"origin\": \"Application\",\
            \"value\": \"$eg_role_name\"\
        }] | {appRoles: .}")
        mgc applications-with-app-id patch --app-id "$webhook_app_id" --output NONE --body "$updated_app_roles"
        app_roles=$(echo $updated_app_roles | jq -r '.appRoles')
    else
        app_role_id=$(echo $app_role | jq -r ".id")
        echo "The role already exists in the Microsoft Entra Application"
    fi

    echo "Microsoft Entra App roles after addition of the new role..."
    echo $app_roles | jq -r ".[].displayName"

    current_user=$(mgc me get)
    current_user_object_id=$(echo $current_user | jq -r ".id")
    current_user_principal_name=$(echo $current_user | jq -r ".userPrincipalName")

    echo "Creating the Microsoft Entra App Role assignment for user: $current_user_principal_name"

    # Creates the user role assignment for the user who will create event subscription
    webhook_sp_id=$(mgc service-principals-with-app-id get --app-id "$webhook_app_id" --select "id" | jq -r ".id") 

    mgc service-principals app-role-assigned-to create --service-principal-id "$webhook_sp_id" --output NONE --body "{\
        \"principalId\": \"$current_user_object_id\",\
        \"resourceId\": \"$webhook_sp_id\",\
        \"appRoleId\": \"$app_role_id\"\
    }\
    "
    
    # Creates the service app role assignment for Event Grid Microsoft Entra Application
    mgc service-principals app-role-assigned-to create --service-principal-id "$webhook_sp_id" --output NONE --body "{\
        \"principalId\": \"$eg_sp_id\",\
        \"resourceId\": \"$webhook_sp_id\",\
        \"appRoleId\": \"$app_role_id\"\
        }\
    "
} > ./entraid.log

echo "KU_APP_ID='$webhook_app_id'" > ../.env.entraid
cat ../.env.entraid