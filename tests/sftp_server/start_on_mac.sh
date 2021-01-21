sudo sh -c 'echo "Subsystem sftp internal-sftp" >> /private/etc/ssh/sshd_config'

sudo launchctl stop com.openssh.sshd
sudo launchctl start com.openssh.sshd

cp tests/sftp_server/id_rsa ~/.ssh/semaphore_cache_key
chmod 0600 ~/.ssh/semaphore_cache_key

ssh-keyscan -p 22 -H localhost >> ~/.ssh/known_hosts

export SEMAPHORE_CACHE_URL=localhost:22
export SEMAPHORE_CACHE_USERNAME=semaphore
