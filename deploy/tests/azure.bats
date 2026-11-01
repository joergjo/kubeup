#!/usr/bin/env bats

load helpers/common

setup() {
    setup_workspace
}

@test "Azure deployment reports successful Event Grid setup" {
    run bash ./azure.sh
    [ "$status" -eq 0 ]
    [[ "$output" == *"Event Grid topic has been created successfully."* ]]
}

@test "Azure deployment propagates Event Grid failure without a success message" {
    export MOCK_EVENTGRID_STATUS=42
    run bash ./azure.sh
    [ "$status" -eq 42 ]
    [[ "$output" == *"Failed to deploy Event Grid subscription."* ]]
    [[ "$output" != *"Event Grid topic has been created successfully."* ]]
}

@test "Azure deployment skips Event Grid when no cluster is configured" {
    export KU_AKS_CLUSTER="" MOCK_EVENTGRID_STATUS=42
    run bash ./azure.sh
    [ "$status" -eq 0 ]
    [[ "$output" == *"Skipping Event Grid topic creation."* ]]
    [[ "$output" != *"Event Grid topic has been created successfully."* ]]
}
