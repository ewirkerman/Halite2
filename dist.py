import argparse
import os
import shutil
import sys
import subprocess

answer = input("Type 'dist':")
if answer != "dist":
    print("Aborting the dist process.")
    sys.exit(1)

bot_path = sys.argv[1]

i = 0
new_path = os.path.join("versions", str(i))
while os.path.exists(new_path):
    i += 1
    new_path = os.path.join("versions", str(i))

shutil.copytree(bot_path, new_path)

subprocess.call("python C:\\Users\\ewirk\\AppData\\Local\\Programs\\Python\\Python35\\lib\\site-packages\\hlt_client\\client.py bot -b %s" % os.path.join(bot_path, "MyBot.zip"))