# Dev toolchain image for riftweave.
# Python is pinned to the same version as the repo-root .toolchain-python-version.
# At run time the repo is mounted at /src. This image holds the interpreter,
# git, and the ruleset Python dependencies. The immutable CI image stays at
# the repository-root Dockerfile.

ARG PYTHON_VERSION=3.13.18
FROM python:${PYTHON_VERSION}-slim-bookworm

ARG PYTHON_VERSION

ENV DEBIAN_FRONTEND=noninteractive \
    PYTHON_VERSION="${PYTHON_VERSION}" \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    PATH="/usr/local/bin:${PATH}"

RUN apt-get update \
    && apt-get install -y --no-install-recommends bash ca-certificates git \
    && rm -rf /var/lib/apt/lists/* /var/tmp/* /tmp/* \
    && python -c "import os, sys; expected = os.environ['PYTHON_VERSION']; actual = sys.version.split()[0]; assert actual == expected, (actual, expected)"

COPY ruleset/requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt \
    && rm -f /tmp/requirements.txt

COPY container/entrypoint.sh /usr/local/bin/entrypoint
RUN chmod 755 /usr/local/bin/entrypoint

WORKDIR /src
ENTRYPOINT ["/usr/local/bin/entrypoint"]
CMD ["python", "ruleset/scripts/validate.py"]
