# Agent image for riftweave: the toolchain image plus Claude Code and uv.
# Built by `devops agent` (scripts/agent.sh) after the toolchain image exists.
# The agent user gets the host's uid and gid so files under /src stay the host user's.
# Its home is a named volume, so logins survive the container and never touch the host's.

ARG BASE_IMAGE=riftweave:python-3.13.18

FROM ghcr.io/astral-sh/uv:0.10.12@sha256:72ab0aeb448090480ccabb99fb5f52b0dc3c71923bffb5e2e26517a1c27b7fec AS uv

FROM ubuntu:24.04 AS node

ARG NODE_VERSION=22.23.4
ARG NODE_SHA256=1805bcd35e6815559bd2eb97db438d3fdb0eb3a5e6a977b7c1badca76337e9ce

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl xz-utils \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL -o /tmp/node.tar.xz \
        "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz" \
    && echo "${NODE_SHA256}  /tmp/node.tar.xz" | sha256sum -c - \
    && mkdir -p /opt/node \
    && tar -xJf /tmp/node.tar.xz -C /opt/node --strip-components=1 \
    && rm -f /tmp/node.tar.xz

FROM ${BASE_IMAGE}

ARG CLAUDE_CODE_VERSION=2.1.312
ARG AGENT_USER_ID=1000
ARG AGENT_GROUP_ID=1000

ENV PATH="/opt/node/bin:${PATH}" \
    DISABLE_AUTOUPDATER=1

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        less \
        openssh-client \
        passwd \
        procps \
        ripgrep \
    && rm -rf /var/lib/apt/lists/* /var/tmp/* /tmp/*

COPY --from=node /opt/node /opt/node
COPY --from=uv /uv /uvx /usr/local/bin/

RUN npm install -g "@anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}" \
    && npm cache clean --force

# The base image may already own this uid (Ubuntu's `ubuntu`, or another default user).
RUN if getent passwd ubuntu >/dev/null; then userdel --remove ubuntu; fi \
    && if id -u "${AGENT_USER_ID}" >/dev/null 2>&1; then \
        userdel "$(id -nu "${AGENT_USER_ID}")"; \
    fi \
    && if ! getent group "${AGENT_GROUP_ID}" >/dev/null; then \
        groupadd --gid "${AGENT_GROUP_ID}" agent; \
    fi \
    && useradd --uid "${AGENT_USER_ID}" --gid "${AGENT_GROUP_ID}" \
        --home-dir /home/agent --create-home --shell /bin/bash agent \
    && mkdir -p /home/agent/.claude \
    && chown -R "${AGENT_USER_ID}:${AGENT_GROUP_ID}" /home/agent \
    && printf '%s\n' 'export PATH="/opt/node/bin:$PATH"' > /etc/profile.d/node-path.sh

ENV HOME=/home/agent \
    CLAUDE_CONFIG_DIR=/home/agent/.claude

USER agent
WORKDIR /src
CMD ["bash"]
