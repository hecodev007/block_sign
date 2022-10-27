FROM node:8.7

WORKDIR /app

RUN mkdir -p /app

COPY . /app

RUN cd /app && \
    npm install && \
    chmod +x get_block_prefix.sh && \
    chmod +x sign_offline_transfer.sh

CMD [ "/bin/bash", "get_block_prefix.sh" ]