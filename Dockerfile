FROM alpine:edge
ADD ./kvgossip /
RUN chmod +x /kvgossip
VOLUME ["/data"]
ENTRYPOINT ["/kvgossip"]
CMD ["--dbpath", "/data/kvgossip.db", "--rootkey", "/data/root_pub.pem", "agent"]
