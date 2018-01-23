import os, sys, shutil, zipfile, glob, subprocess

returncode = subprocess.call("go build MyBot.go", shell=True)
if returncode != 0:
    raise Exception("Bot did not build correctly")

BOT_NAME = None

with open(os.path.join("src","app.go")) as infile:
    for line in infile.readlines():
        if "NAME    = " in line:
            BOT_NAME = line.split("=")[-1].strip(" \"\n")
            break

def zipdir(path, ziph):
    # ziph is zipfile handle
    for root, dirs, files in os.walk(path):
        for file in files:
            ziph.write(os.path.join(root, file))

def onerror(func, path, exc_info):
    """
    Error handler for ``shutil.rmtree``.

    If the error is due to an access error (read only file)
    it attempts to add write permission and then retries.

    If the error is for another reason it re-raises the error.

    Usage : ``shutil.rmtree(path, onerror=onerror)``
    """
    import stat
    if not os.access(path, os.W_OK):
        # Is the error an access error ?
        os.chmod(path, stat.S_IWUSR)
        func(path)
    else:
        raise
	  
def find_and_replace(filename, replacements={}):
    """
    :type filename str
    :type replacements dict
    :param filename:
    :param replacements:
    :return:
    """
    lines = []
    with open(filename) as infile:
        for line in infile.readlines():
            for src, target in replacements.items():
                line = line.replace(src, target)
            lines.append(line)
    with open(filename, 'w') as outfile:
        for line in lines:
            outfile.write(line)
		
path = os.path.join("bots",BOT_NAME)
	  
import filecmp
new_src = False


if os.path.exists(os.path.join(path, "src")):
	for f in os.listdir(os.path.join(path, "src")):
		if ".go" in f and not filecmp.cmp(os.path.join(path, "src", f), os.path.join("src", f), shallow=False):
			new_src = True
			break
else:
	new_src = True
	
if new_src:
	if os.path.exists(path):
	    shutil.rmtree(path, onerror=onerror)
	os.makedirs(path)

	shutil.copyfile("MyBot.go", os.path.join(path, "MyBot.go"))
	shutil.copyfile("install.sh", os.path.join(path, "install.sh"))
	shutil.copytree("src", os.path.join(path, "src"))

	zipPath = os.path.join(path, "MyBot.zip")
	with zipfile.ZipFile(zipPath, 'w') as myzip:
	    myzip.write("MyBot.go")
	    # myzip.write("install.sh")
	    zipdir("src", myzip)

	


# for filename in os.listdir(os.path.join(path, "src")):
#     if filename.endswith(".go"):
#         find_and_replace(os.path.join(path, "src", filename), {"game.Log(":"//game.Log","g.Log":"//g.Log"})
# find_and_replace(os.path.join(path, "MyBot.go"), {"game.Log":"//game.Log","g.Log":"//g.Log"})



print(BOT_NAME)
