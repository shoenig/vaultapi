curl -o /tmp/vault.zip  https://releases.hashicorp.com/vault/0.7.3/vault_0.7.3_linux_amd64.zip
unzip /tmp/vault.zip -d /tmp
tree /tmp
touch /tmp/dev-vault.token
chmod 0666 /tmp/dev-vault.token