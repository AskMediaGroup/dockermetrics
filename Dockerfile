FROM debian:latest
MAINTAINER Richard Marshall <rgm@linux.com>

ADD dockermetrics /

ENTRYPOINT ["/dockermetrics"]
