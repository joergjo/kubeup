#!/usr/bin/env bats

load helpers/common

setup() {
    setup_workspace
}

@test "Entra setup reuses saved configuration without any Azure CLI call" {
    printf "KU_APP_ID='%s'\n" "$APP_ID" > "$TEST_ROOT/.env.entraid"
    printf '#!/bin/bash\necho unexpected >> "$CALL_LOG"\nexit 42\n' > "$TEST_ROOT/bin/az"
    for invocation in 1 2; do
        run bash ./entraid.sh
        [ "$status" -eq 0 ]
        [ "$output" = "KU_APP_ID='$APP_ID'" ]
        assert_saved_config
    done
    [ ! -e "$CALL_LOG" ]
}

@test "Entra setup creates objects, preserves roles and normalizes macOS UUIDs" {
    export MOCK_MODE=new
    run bash ./entraid.sh
    [ "$status" -eq 0 ]
    assert_saved_config
    jq -es --arg user "$USER_ID" --arg grid "$EVENTGRID_SP" \
        '[.[] | select(.method == "POST" and (.url | endswith("/appRoleAssignedTo"))) | .body.principalId] |
        sort == ([$user,$grid] | sort)' "$CALL_LOG"
}

@test "Entra setup reuses an existing application, principals and assignments" {
    export MOCK_MODE=existing
    run bash ./entraid.sh "$APP_ID"
    [ "$status" -eq 0 ]
    assert_saved_config
    assert_no_mutations
}

@test "Entra setup follows assignment pagination without duplicating assignments" {
    export MOCK_MODE=paginated
    run bash ./entraid.sh "$APP_ID"
    [ "$status" -eq 0 ]
    assert_saved_config
    assert_no_mutations
    jq -es 'any(.[]; .url | endswith("?page=2"))' "$CALL_LOG"
}

@test "Entra setup retries delayed role propagation" {
    export MOCK_MODE=propagation
    run bash ./entraid.sh
    [ "$status" -eq 0 ]
    assert_saved_config
    jq -es --arg sp "$WEBHOOK_SP" '[.[] | select(.url | endswith("/" + $sp))] | length == 2' "$CALL_LOG"
}

@test "Entra setup fails explicitly and never saves configuration after errors" {
    for mode in unknown-app no-login non-user forbidden missing-id malformed-lookup create-failure principal-missing-id patch-failure assignment-failure no-role; do
        export MOCK_MODE="$mode"
        rm -f "$CALL_LOG" ./entraid.log
        if [[ "$mode" == unknown-app ]]; then
            run bash ./entraid.sh "$APP_ID"
        else
            run bash ./entraid.sh
        fi
        if [[ "$status" -eq 0 || -z "$output" ]]; then
            echo "Expected explicit failure for $mode: $output"
            return 1
        fi
        [ ! -e "$TEST_ROOT/.env.entraid" ]
        if [[ "$mode" == forbidden ]]; then
            assert_no_mutations_after_lookup=$(jq -s '[.[] | select(.url | contains("servicePrincipals")) | select(.method == "POST")] | length' "$CALL_LOG")
            [ "$assert_no_mutations_after_lookup" -eq 0 ]
        fi
        if [[ "$mode" == no-role ]]; then
            jq -es --arg sp "$WEBHOOK_SP" '[.[] | select(.url | endswith("/" + $sp))] | length == 5' "$CALL_LOG"
        fi
    done
}
