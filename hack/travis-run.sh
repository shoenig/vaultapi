#!/bin/bash

# Start vault in dev mode, which puts a root token in ~/.vault-token
/tmp/vault server -dev &
sleep 3
ps -ef | grep vault

# Copy that root token to /tmp/root.token
cp ~/.vault-token /tmp/root.token
echo "vault root token: $(cat /tmp/root.token)"

# Set environment to talk to vault with root token
export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN=$(head -n 1 /tmp/root.token)

# Create an example policy with permissions to
# - use secret/my/stuff to read/create/update/list keys
# - use auth/token/lookup-self to look itself up
# - use auth/token/renew-self to renew itself (note, also requires a role)
# - cannot do anything with sys/
cat << EOF> /tmp/my_policy1
path "sys/*" {
    policy = "deny"
}

path "secret/my/stuff/*" {
    capabilities = ["read", "create", "update", "list"]
}

path "auth/token/lookup-self" {
    policy = "read"
}

path "auth/token/renew-self" {
    policy = "write"
}
EOF

# Write my_policy1 into vault
/tmp/vault policy-write my_policy1 /tmp/my_policy1 || exit 1

# Create a non-renewable token based on my_policy1
my_token1=$(/tmp/vault token-create -policy=my_policy1 -orphan=true -format=json | jq -r .auth.client_token)
echo ${my_token1} > /tmp/t1.token
echo "t1 token: $(cat /tmp/t1.token)"

# Create an example role which enables renewable tokens
/tmp/vault write auth/token/roles/my_role1 allowed_policies=my_policy1 orphan=true period=10800 renewable=true
# Create a renewable token based on my_role which
# incorporates my_policy1 and allows renewable tokens
# with a period of 3 hours (10800 seconds).
my_token2=$(/tmp/vault token-create -role=my_role1 -policy=my_policy1 -orphan=true -format=json | jq -r .auth.client_token)
echo ${my_token2} > /tmp/t2.token
echo "t2 token: $(cat /tmp/t2.token)"