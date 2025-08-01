# Options to explore

List of options:

- pm2 for golang
- https://medium.com/@a-yon_n/use-pm2-to-monitor-golang-programs-01af6ead02af
- systemd is too unwieldy
- https://medium.com/@a-yon_n/use-pm2-to-monitor-golang-programs-01af6ead02af
- https://github.com/ayonli/ngrpc?source=post_page-----01af6ead02af---------------------------------------
- systemd command line
- https://pm2.io/docs/enterprise/collector/go/
- supervisord
- https://github.com/takama/daemon
- https://github.com/ShinyTrinkets/overseer
- https://github.com/topics/process-manager?l=go
- https://awesome-go.com/
- https://github.com/restuwahyu13/golang-pm2
- https://github.com/cschleiden/go-workflows
- temporal, ingest
- nohup?
- start-stop-daemon
- https://github.com/ochinchina/supervisord
- https://github.com/DarthSim/hivemind

Keywords:

- Durable workflows
- service runner
- program manager

Good options:

- [pmgo](https://github.com/struCoder/pmgo). Looks good, well mantained
- Maybe <https://github.com/immortal/immortal?tab=readme-ov-file>

# Considered and discarded because not mantained

daemontools
supervisord

# Top options

- Keep systemd
  - This seems like the best option for now. In particular, the thing that was maybe messing with me was having a hard to update central makefile, rather than one makefile for each service.
- Use pm2 for golang
  - <https://github.com/restuwahyu13/golang-pm2>
