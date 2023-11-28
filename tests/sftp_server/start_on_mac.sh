sudo dscl . -create /Users/sftp
sudo dscl . -create /Users/sftp UserShell /bin/bash
sudo dscl . -create /Users/sftp RealName "SFTP"
sudo dscl . -create /Users/sftp UniqueID "1010"
sudo dscl . -create /Users/sftp PrimaryGroupID 80
sudo dscl . -create /Users/sftp NFSHomeDirectory /Users/sftp

sudo mkdir -p /Users/sftp

sudo sh -c 'echo "Subsystem sftp internal-sftp" > /private/etc/ssh/sshd_config'
sudo sh -c 'echo "AuthorizedKeysFile  /etc/ssh/authorized_keys/%u" >> /private/etc/ssh/sshd_config'
sudo sh -c 'echo "PubkeyAcceptedAlgorithms +ssh-rsa" >> /private/etc/ssh/sshd_config'
sudo sh -c 'echo "PubkeyAcceptedKeyTypes=+ssh-rsa" >> /private/etc/ssh/sshd_config'

sudo mkdir -p /private/etc/ssh/authorized_keys

sudo sh -c 'cat tests/sftp_server/id_rsa.pub >> /private/etc/ssh/authorized_keys/sftp'
sudo chown -R sftp /Users/sftp/
sudo chown -R sftp /private/etc/ssh/authorized_keys/sftp
chmod 0600 /private/etc/ssh/authorized_keys/sftp

sudo launchctl stop com.openssh.sshd
sudo launchctl start com.openssh.sshd

cp tests/sftp_server/id_rsa ~/.ssh/semaphore_cache_key
chmod 0600 ~/.ssh/semaphore_cache_key

ssh-keyscan -p 22 -H localhost >> ~/.ssh/known_hosts

export SEMAPHORE_CACHE_URL=localhost:22
export SEMAPHORE_CACHE_USERNAME=sftp
