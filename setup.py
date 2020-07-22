import os
from setuptools import setup

setup(
    name="aws-fuzzy",
    version=open(os.path.join('aws_fuzzy', 'VERSION')).read().strip(),
    packages=["aws_fuzzy", "aws_fuzzy.commands"],
    package_data={'': [
        'VERSION',
    ]},
    install_requires=[
        "click", "boto3>=1.12", "botocore>=1.15", "pygments", "iterfzf",
        "pyvis"
    ],
    python_requires='>=3',
    entry_points="""
        [console_scripts]
        aws-fuzzy=aws_fuzzy.cli:cli
    """,
)
