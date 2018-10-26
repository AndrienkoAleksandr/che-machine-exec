#
# Copyright (c) 2012-2018 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#

FROM golang:1.10.3 as builder
WORKDIR /go/src/github.com/eclipse/che-machine-exec/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -a -installsuffix cgo -o che-machine-exec .

FROM registry.centos.org/centos:7
RUN yum -y update && yum -y install mc nano
COPY --from=builder /go/src/github.com/eclipse/che-machine-exec/che-machine-exec /usr/local/bin
RUN touch /usr/local/bin/restore && chmod 777 /usr/local/bin/restore
RUN printf true >> /usr/local/bin/restore
ENTRYPOINT ["che-machine-exec"]
