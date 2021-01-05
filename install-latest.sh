#!/bin/bash

set -e

pushd /tmp/

curl -sSL https://api.github.com/repos/instadeepai/gcp-gpu-metrics/releases/latest \
| grep "browser_download_url.*gcp-gpu-metrics_Linux_x86_64\.tar\.gz" \
| cut -d ":" -f 2,3 \
| tr -d \" \
| wget -qi -

tar -xzf './gcp-gpu-metrics_Linux_x86_64.tar.gz'

chmod +x gcp-gpu-metrics

mv gcp-gpu-metrics /usr/local/bin/

popd

echo -e "✅ gcp-gpu-metrics binary installed ✅"