#coding:utf-8
import os
import shutil

if os.path.exists("docs/_sources"):
    shutil.rmtree("docs/sources")
    shutil.move("docs/_sources","docs/sources/")

if os.path.exists("docs/_static"):
    shutil.rmtree("docs/static/")
    shutil.move("docs/_static","docs/static/")

for dirpath, dirnames, filenames in os.walk("docs"):
    for filename in filenames:
        if filename.startswith("."):
            continue
        fname  = os.path.join(dirpath,filename)
        print "rename ",fname
        newf = open(fname,'r').read().replace("_static/","static/").replace("_sources/","sources/")
        open(fname, 'wb').write(newf)
