import os

import pytest
import requests


@pytest.fixture(scope="session")
def base_url():
    port = os.environ.get("JUNO_REST_PORT", "6000")
    return f"http://localhost:{port}"
