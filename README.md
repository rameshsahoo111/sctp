# sctp
Sample Go SCTP Application 

# Image build steps 
`podman build -t sctp -f Containerfile .`

## Test 
```bash
# /app/sctp-server -port 5000
OR
# /app/sctp-server
SCTP server listening on 0.0.0.0:5000 (via syscall)
Received from 10.88.0.14:50908: Ping from native Go SCTP Client

# app/sctp-client -ip 10.88.0.14 -port 5000 
Connecting to 10.88.0.14:5000...
Connected successfully!
Sent: Ping from native Go SCTP Client
Server Reply: Hello from SCTP Server! My Hostname: sctp, My IP: 10.88.0.14:5000 | Your Client IP: 10.88.0.14:50908
```


