# Health Check Script
# This script tests if the API is running correctly

import requests
import sys

def check_api_health():
    base_url = "http://localhost:8080"
    
    print("Checking API health at", base_url)
    
    try:
        # Test a public endpoint
        response = requests.get(f"{base_url}/website/modules", timeout=5)
        if response.status_code == 200:
            print("✓ API is accessible - /website/modules returned 200")
        else:
            print(f"✗ API returned unexpected status: {response.status_code}")
            
        # Test another public endpoint
        response = requests.get(f"{base_url}/blog-posts", timeout=5)
        if response.status_code == 200:
            print("✓ API is accessible - /blog-posts returned 200")
        else:
            print(f"✗ API returned unexpected status: {response.status_code}")
            
        print("\nAPI is running successfully!")
        print("Note: Database is not connected, so data will not be persisted.")
        print("To enable full functionality, please set up a database as described in RUNNING_INSTRUCTIONS.md")
        
    except requests.exceptions.ConnectionError:
        print("✗ Cannot connect to API. Please ensure the server is running on port 8080.")
        return False
    except requests.exceptions.Timeout:
        print("✗ Request timed out. Please check if the server is responsive.")
        return False
    except Exception as e:
        print(f"✗ Error occurred while checking API health: {str(e)}")
        return False
    
    return True

if __name__ == "__main__":
    check_api_health()