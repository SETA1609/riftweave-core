# Dev toolchain image for riftweave, on the devops-tools base image (tools/devops/README.md).
# `devops-tools container` builds the base and passes it as BASE_IMAGE. This image adds only
# the pinned Python and the ruleset's Python dependencies, in /opt/riftweave-venv, first on
# PATH. The version must equal the repo-root .toolchain-python-version; the runner checks it.
# The immutable CI image stays at the repository-root Dockerfile.

ARG BASE_IMAGE
FROM ${BASE_IMAGE}

ARG PYTHON_VERSION=3.13.12

ENV VIRTUAL_ENV=/opt/riftweave-venv \
    PATH="/opt/riftweave-venv/bin:${PATH}" \
    PYTHONDONTWRITEBYTECODE=1

COPY ruleset/requirements.txt /tmp/requirements.txt
RUN uv python install "${PYTHON_VERSION}" \
    && uv venv --python "${PYTHON_VERSION}" /opt/riftweave-venv \
    && uv pip install --no-cache --python /opt/riftweave-venv/bin/python -r /tmp/requirements.txt \
    && rm -f /tmp/requirements.txt \
    && chmod -R a+rX /opt/uv /opt/riftweave-venv \
    && /opt/riftweave-venv/bin/python -c "import sys; v = sys.version.split()[0]; assert v == '${PYTHON_VERSION}', v"

CMD ["python", "ruleset/scripts/validate.py"]
