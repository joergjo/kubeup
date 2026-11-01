setup_workspace() {
    export TEST_ROOT="$BATS_TEST_TMPDIR/workspace"
    export CALL_LOG="$BATS_TEST_TMPDIR/calls.jsonl"
    mkdir -p "$TEST_ROOT/deploy" "$TEST_ROOT/bin"
    cp "$BATS_TEST_DIRNAME/../azure.sh" "$BATS_TEST_DIRNAME/../entraid.sh" "$TEST_ROOT/deploy/"
    cp "$BATS_TEST_DIRNAME/helpers/az" "$TEST_ROOT/bin/az"
    chmod +x "$TEST_ROOT/bin/az"
    printf '#!/bin/bash\necho ABCDEFAB-ABCD-ABCD-ABCD-ABCDEFABCDEF\n' > "$TEST_ROOT/bin/uuidgen"
    printf '#!/bin/bash\nexit 0\n' > "$TEST_ROOT/bin/sleep"
    chmod +x "$TEST_ROOT/bin/uuidgen" "$TEST_ROOT/bin/sleep"
    export PATH="$TEST_ROOT/bin:$PATH"
    export AZURE_CONFIG_DIR="$TEST_ROOT/azure"
    export APP_ID=11111111-1111-1111-1111-111111111111
    export OBJECT_ID=22222222-2222-2222-2222-222222222222
    export WEBHOOK_SP=33333333-3333-3333-3333-333333333333
    export EVENTGRID_SP=44444444-4444-4444-4444-444444444444
    export USER_ID=55555555-5555-5555-5555-555555555555
    export ROLE_ID=abcdefab-abcd-abcd-abcd-abcdefabcdef
    export KU_RESOURCE_GROUP=webhook-rg KU_AKS_RESOURCE_GROUP=aks-rg KU_AKS_CLUSTER=cluster
    unset KU_LOCATION KU_IMAGE KU_SENDGRID_APIKEY KU_EMAIL_FROM KU_EMAIL_TO KU_EMAIL_SUBJECT
    unset KU_SMTP_HOST KU_SMTP_PORT KU_SMTP_USERNAME KU_SMTP_PASSWORD KU_SECRET_1 KU_SECRET_2 KU_APP_ID
    cd "$TEST_ROOT/deploy" || return
}

assert_saved_config() {
    [ "$(cat "$TEST_ROOT/.env.entraid")" = "KU_APP_ID='$APP_ID'" ]
}

assert_no_mutations() {
    jq -es 'all(.[]; .method != "POST" and .method != "PATCH")' "$CALL_LOG"
}
