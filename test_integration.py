#!/usr/bin/env python3
"""
Quick test script for Drawy Claude Code integration
"""

import subprocess
import sys

def run_cmd(cmd):
    """Run a command and return success"""
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        if result.returncode == 0:
            print(f"✓ {cmd}")
            if result.stdout.strip():
                print(f"  Output: {result.stdout.strip()}")
        else:
            print(f"✗ {cmd}")
            print(f"  Error: {result.stderr.strip()}")
        return result.returncode == 0
    except Exception as e:
        print(f"✗ {cmd} - Exception: {e}")
        return False

def test_integration():
    """Test the Drawy integration"""
    print("🧪 Testing Drawy Claude Code Integration")
    print("=" * 50)

    # Test CLI availability
    print("\n1. Testing CLI availability...")
    if not run_cmd("python cli/main.py --help | head -5"):
        return False

    # Test initialization
    print("\n2. Testing board initialization...")
    if not run_cmd('python cli/main.py init --topic "Integration Test"'):
        return False

    # Test topic connection
    print("\n3. Testing topic connection...")
    if not run_cmd('python cli/main.py expand --topic "Test Connection"'):
        return False

    # Test expansion
    print("\n4. Testing expansion...")
    run_cmd("python cli/main.py expand --steps 1")  # May or may not add concepts

    # Test export
    print("\n5. Testing export...")
    if not run_cmd("python cli/main.py export --out integration-test.svg"):
        return False

    # Check file sizes
    print("\n6. Checking file sizes...")
    run_cmd("ls -la infinite.*.json integration-test.svg")

    print("\n🎉 Integration test complete!")
    print("\n📖 Next: Follow CLAUDE_CODE_INTEGRATION.md to install in Claude Code")
    print("🌐 Run: python serve_canvas.py to view the infinite canvas")
    return True

if __name__ == "__main__":
    success = test_integration()
    sys.exit(0 if success else 1)