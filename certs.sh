#!/usr/bin/env bash

# Adapted from https://gist.github.com/fntlnz/cf14feb5a46b2eda428e000157447309

rootCAKey=root_ca.key
rootCACrt=root_ca.crt
domainKey=domain.key
domainCrs=domain.crs
domainCrt=domain.crt

rm $rootCAKey
rm $rootCACrt
rm $domainKey
rm $domainCrs
rm $domainCrt

# {{- $altNames := list ( printf "agent.%s" .Release.Namespace ) ( printf "agent.%s.svc" .Release.Namespace ) -}}
# {{- $ca := genCA (printf "%s-ca" .Chart.Name) 3650 -}}
# {{- $cert := genSignedCert .Chart.Name nil $altNames 3650 $ca -}}

# 1. Generate the root CA key
openssl genrsa -out $rootCAKey 4096

# 2. Generate the self-signed root CA. Valid for 5years(1825 days)
# -subj "/emailAddress=tim@traceable.ai/C=US/ST=California/L=San Francisco/O=Traceable AI, Inc./OU=Engineering/CN=agent.traceableai" \
openssl req -x509 -new -nodes -sha256 -key $rootCAKey -days 1825 \
    -subj "/CN=traceable-agent-ca" \
    -out $rootCACrt

# 3. Generate the Certificate key
openssl genrsa -out $domainKey 4096

# 4. Generate the Certificate Request. Valid for 5years(1825 days)
#
#
# -subj "/emailAddress=tim@traceable.ai/C=US/ST=California/L=San Francisco/O=Traceable AI, Inc./OU=Engineering/CN=agent.traceableai" \
# a printf with more alternative names
# <(printf "\n[SAN]\nsubjectAltName=DNS.1:agent.traceableai,DNS.2:agent.traceableai.svc,DNS.3:localhost,DNS.4:0.0.0.0,DNS.5:host.docker.internal,DNS.6:127.0.0.1")) \
openssl req -new -sha256 -key $domainKey \
    -subj "/CN=traceable-agent" \
    -reqexts SAN \
    -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\n[SAN]\nsubjectAltName=DNS.1:extprocserver,DNS.2:agent.traceableai.svc")) \
    -out $domainCrs

# 4b. Quick verify
openssl req -in $domainCrs -noout -text

# There is a bug in x509 command which does not allow the subjectAltName to be copied over from the crs. So we use the
# -extfile cmd line option. See https://gist.github.com/fntlnz/cf14feb5a46b2eda428e000157447309#gistcomment-3034183
#
# -extfile <(printf "subjectAltName=DNS.1:agent.traceableai,DNS.2:agent.traceableai.svc,DNS.3:localhost,DNS.4:0.0.0.0,DNS.5:host.docker.internal,DNS.6:127.0.0.1")
# 5. Generate the Certificate using the root CA
openssl x509 -req -in $domainCrs -CA $rootCACrt -CAkey $rootCAKey -CAcreateserial -days 1825 -sha256 -out $domainCrt \
  -extfile <(printf "subjectAltName=DNS.1:extprocserver,DNS.2:agent.traceableai.svc")

# 5b. Quick verify
openssl x509 -in $domainCrt -text -noout
