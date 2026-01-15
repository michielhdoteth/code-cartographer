"""
Framework Detector Module

Detects frameworks and libraries used in codebases.
"""

from src.detector.framework import FrameworkDetector, FrameworkInfo, ProjectFrameworks, detect_frameworks

__all__ = ["FrameworkDetector", "FrameworkInfo", "ProjectFrameworks", "detect_frameworks"]
