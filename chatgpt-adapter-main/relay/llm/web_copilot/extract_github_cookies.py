#!/usr/bin/env python3
"""
GitHub Cookie Extractor
Extract cookies from browser for GitHub Copilot authentication.
"""

import os
import sys
import json
import sqlite3
import shutil
import tempfile
import argparse
from pathlib import Path
from datetime import datetime, timedelta

def get_firefox_profile_dirs():
    """Find Firefox profile directories on the current system."""
    profiles = []
    
    if sys.platform.startswith('win'):
        base_path = os.path.join(os.environ.get('APPDATA', ''), 'Mozilla', 'Firefox', 'Profiles')
    elif sys.platform.startswith('darwin'):
        base_path = os.path.expanduser('~/Library/Application Support/Firefox/Profiles')
    else:  # Linux and others
        base_path = os.path.expanduser('~/.mozilla/firefox')
    
    if not os.path.exists(base_path):
        return profiles
    
    # Handle direct profiles directory
    if os.path.isdir(base_path):
        for item in os.listdir(base_path):
            profile_path = os.path.join(base_path, item)
            if os.path.isdir(profile_path) and item.endswith('.default') or '.default-' in item:
                profiles.append((item, profile_path))
    
    # Handle profiles.ini approach
    profiles_ini = os.path.join(os.path.dirname(base_path), 'profiles.ini')
    if os.path.exists(profiles_ini):
        with open(profiles_ini, 'r') as f:
            current_profile = None
            current_path = None
            
            for line in f:
                line = line.strip()
                if line.startswith('[Profile'):
                    if current_profile and current_path:
                        profiles.append((current_profile, current_path))
                    current_profile = None
                    current_path = None
                elif '=' in line:
                    key, value = line.split('=', 1)
                    if key == 'Name':
                        current_profile = value
                    elif key == 'Path':
                        if os.path.isabs(value):
                            current_path = value
                        else:
                            current_path = os.path.join(os.path.dirname(profiles_ini), value)
            
            if current_profile and current_path:
                profiles.append((current_profile, current_path))
    
    return profiles

def copy_cookie_db(profile_path):
    """Create a temporary copy of the cookies database to avoid lock issues."""
    cookies_db = os.path.join(profile_path, 'cookies.sqlite')
    
    if not os.path.exists(cookies_db):
        return None
    
    temp_dir = tempfile.mkdtemp()
    temp_db = os.path.join(temp_dir, 'cookies.sqlite')
    
    try:
        shutil.copy2(cookies_db, temp_db)
        return temp_db
    except Exception as e:
        print(f"Error copying cookies database: {e}")
        return None

def extract_cookies(db_path, domain_filter=None):
    """Extract cookies from the Firefox SQLite database."""
    cookies = {}
    
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        query = '''
            SELECT host, name, value, path, expiry, isSecure, isHttpOnly, sameSite
            FROM moz_cookies
        '''
        
        if domain_filter:
            query += " WHERE host LIKE ?"
            cursor.execute(query, (f"%{domain_filter}%",))
        else:
            cursor.execute(query)
        
        for row in cursor.fetchall():
            host, name, value, path, expiry, is_secure, is_http_only, same_site = row
            
            # Add cookie to the map
            cookies[name] = value
        
        conn.close()
        return cookies
    
    except Exception as e:
        print(f"Error extracting cookies: {e}")
        return {}

def save_as_json(cookies, output_path):
    """Save cookies in JSON format."""
    with open(output_path, 'w') as f:
        json.dump(cookies, f, indent=2)
    
    print(f"Saved {len(cookies)} cookies to {output_path}")

def main():
    parser = argparse.ArgumentParser(description='Extract GitHub cookies from Firefox browser')
    parser.add_argument('-o', '--output', default='github_cookies.json', help='Output file path')
    parser.add_argument('-p', '--profile', help='Specific profile name to use')
    parser.add_argument('-l', '--list-profiles', action='store_true', help='List available Firefox profiles')
    
    args = parser.parse_args()
    
    profiles = get_firefox_profile_dirs()
    
    if not profiles:
        print("No Firefox profiles found.")
        return
    
    if args.list_profiles:
        print("Available Firefox profiles:")
        for i, (name, path) in enumerate(profiles):
            print(f"{i+1}. {name} ({path})")
        return
    
    # Select profile
    selected_profile = None
    
    if args.profile:
        for name, path in profiles:
            if args.profile in name:
                selected_profile = path
                break
        
        if not selected_profile:
            print(f"Profile '{args.profile}' not found.")
            return
    elif len(profiles) == 1:
        selected_profile = profiles[0][1]
    else:
        print("Available Firefox profiles:")
        for i, (name, path) in enumerate(profiles):
            print(f"{i+1}. {name} ({path})")
        
        choice = input("Select profile (number): ")
        try:
            index = int(choice) - 1
            if 0 <= index < len(profiles):
                selected_profile = profiles[index][1]
            else:
                print("Invalid selection.")
                return
        except ValueError:
            print("Invalid input.")
            return
    
    # Create a temporary copy of the cookies database
    temp_db = copy_cookie_db(selected_profile)
    
    if not temp_db:
        print("Could not access cookies database.")
        return
    
    try:
        # Extract cookies for GitHub
        cookies = extract_cookies(temp_db, "github.com")
        
        if not cookies:
            print("No cookies found for GitHub.")
            return
        
        # Save cookies to JSON file
        save_as_json(cookies, args.output)
        
        print("\nTo use these cookies with the web_copilot adapter, include them in your API request:")
        print('Authorization: Bearer ' + json.dumps(cookies))
        
    finally:
        # Clean up the temporary directory
        temp_dir = os.path.dirname(temp_db)
        if os.path.exists(temp_dir):
            shutil.rmtree(temp_dir)

if __name__ == "__main__":
    main()

