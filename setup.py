from setuptools import setup, find_packages

setup(
    name="code-map",
    version="0.0.3",
    packages=find_packages(),
    entry_points={
        'console_scripts': [
            'code-map=cli.main:main',
        ],
    },
    install_requires=[
        'dataclasses-json',
    ],
    author="Code Map Team",
    author_email="contact@codemap.dev",
    description="Code Map - Interactive knowledge graph visualization with infinite canvas",
    long_description=open('README.md').read(),
    long_description_content_type="text/markdown",
    url="https://github.com/code-map/code-map",
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: GNU General Public License v3 (GPLv3)",
        "Operating System :: OS Independent",
    ],
    python_requires='>=3.6',
)
