from setuptools import setup

setup(
    name="aws-fuzzy",
    version="1.0",
    packages=["aws_fuzzy", "aws_fuzzy.commands"],
    include_package_data=True,
    install_requires=["click", "boto3", "pygments"],
    entry_points="""
        [console_scripts]
        aws-fuzzy=aws_fuzzy.cli:cli
    """,
)
