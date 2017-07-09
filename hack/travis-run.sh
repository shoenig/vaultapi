/tmp/vault server -dev &
sleep 3
ps -ef | grep vault
cp ~/.vault-token /tmp/dev-vault.token
echo "vault token: $(cat /tmp/dev-vault.token)"