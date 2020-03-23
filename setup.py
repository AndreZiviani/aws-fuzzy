from setuptools import setup

setup(
    name="aws-fuzzy",
    version="0.0.1",
    packages=["aws_fuzzy", "aws_fuzzy.commands"],
    include_package_data=True,
    install_requires=["click", "boto3", "pygments", "iterfzf"],
    entry_points="""
        [console_scripts]
        aws-fuzzy=aws_fuzzy.cli:cli
    """,
)
