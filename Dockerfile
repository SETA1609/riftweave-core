FROM python:3.12-slim

WORKDIR /app

# Install Python dependencies first (better layer caching)
COPY ruleset/requirements.txt ./ruleset/requirements.txt
RUN pip install --no-cache-dir -r ruleset/requirements.txt

# Copy ruleset (schemas, data, and validation script)
COPY ruleset/ ./ruleset/

# Create non-root user for running validation
RUN useradd --create-home --uid 10001 validator \
    && chown -R validator:validator /app
USER validator

# Default command runs the schema validator
# Override with: docker run --rm riftweave-validate
ENTRYPOINT ["python", "ruleset/scripts/validate.py"]
CMD []
