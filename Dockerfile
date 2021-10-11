FROM alpine:latest

ADD motioneye-snapshotter /app/
CMD ["/app/motioneye-snapshotter", "--config", "/config/config.yaml"]
