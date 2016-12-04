FROM alpine:edge
ADD ./kvgossip ./dumb-init /
RUN chmod +x /kvgossip /dumb-init
VOLUME ["/data"]
ENTRYPOINT ["/dumb-init"]
CMD ["/kvgossip", "--dbpath", "/data/kvgossip.db", "--rootkey", "/data/root_pub.pem", "agent"]
