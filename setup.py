from setuptools import setup
import os

setup(
    name="aws-fuzzy",
    version_file=open(os.path.join('aws_fuzzy', 'VERSION')),
    version=version_file.read().strip(),
    packages=["aws_fuzzy", "aws_fuzzy.commands"],
    include_package_data=True,
    install_requires=["click", "boto3", "pygments", "iterfzf"],
    python_requires='>=3',
    entry_points="""
        [console_scripts]
        aws-fuzzy=aws_fuzzy.cli:cli
    """,
)
