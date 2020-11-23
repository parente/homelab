FROM python:3
COPY requirements.txt /usr/local/src/
RUN pip install --no-cache-dir -r /usr/local/src/requirements.txt
COPY cfsync.py /usr/local/src/
USER nobody
CMD ["python", "/usr/local/src/cfsync.py"]
