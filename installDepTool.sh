#!/bin/bash
#
# Copyright (c) 2012-2017 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#

echo '===>Install dep tool<==='

DEP_DOWNLOAD_URL=https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64
curl -fsSL ${DEP_DOWNLOAD_URL} -o /usr/bin/dep
chmod +x /usr/bin/dep

echo '===>Dep tool successfully installed<==='
