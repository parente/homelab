FROM --platform=$TARGETPLATFORM python:3.9

LABEL org.opencontainers.image.source https://github.com/parente/homelab

COPY requirements.txt /usr/local/src/
RUN pip install --no-cache-dir -r /usr/local/src/requirements.txt
COPY cfsync.py /usr/local/src/
USER nobody
CMD ["python", "/usr/local/src/cfsync.py"]
