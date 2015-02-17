# dictfile.py
#
# Copyright 2009 Kristoffer Gronlund <kristoffer.gronlund@purplescout.se>

""" Dictionary File

Implements an iterable file format that handles the
RADIUS $INCLUDE directives behind the scene.
"""

import os
import six


class _Node(object):
    """Dictionary file node

    A single dictionary file.
    """
    __slots__ = ('name', 'lines', 'current', 'length', 'dir')

    def __init__(self, fd, name, parentdir):
        self.lines = fd.readlines()
        self.length = len(self.lines)
        self.current = 0
        self.name = os.path.basename(name)
        path = os.path.dirname(name)
        if os.path.isabs(path):
            self.dir = path
        else:
            self.dir = os.path.join(parentdir, path)

    def Next(self):
        if self.current >= self.length:
            return None
        self.current += 1
        return self.lines[self.current - 1]


class DictFile(object):
    """Dictionary file class

    An iterable file type that handles $INCLUDE
    directives internally.
    """
    __slots__ = ('stack')

    def __init__(self, fil):
        """
        @param fil: a dictionary file to parse
        @type fil: string or file
        """
        self.stack = []
        self.__ReadNode(fil)

    def __ReadNode(self, fil):
        node = None
        parentdir = self.__CurDir()
        if isinstance(fil, six.string_types):
            fname = None
            if os.path.isabs(fil):
                fname = fil
            else:
                fname = os.path.join(parentdir, fil)
            fd = open(fname, "rt")
            node = _Node(fd, fil, parentdir)
            fd.close()
        else:
            node = _Node(fil, '', parentdir)
        self.stack.append(node)

    def __CurDir(self):
        if self.stack:
            return self.stack[-1].dir
        else:
            return os.path.realpath(os.curdir)

    def __GetInclude(self, line):
        line = line.split("#", 1)[0].strip()
        tokens = line.split()
        if tokens and tokens[0].upper() == '$INCLUDE':
            return " ".join(tokens[1:])
        else:
            return None

    def Line(self):
        """Returns line number of current file
        """
        if self.stack:
            return self.stack[-1].current
        else:
            return -1

    def File(self):
        """Returns name of current file
        """
        if self.stack:
            return self.stack[-1].name
        else:
            return ''

    def __iter__(self):
        return self

    def __next__(self):
        while self.stack:
            line = self.stack[-1].Next()
            if line == None:
                self.stack.pop()
            else:
                inc = self.__GetInclude(line)
                if inc:
                    self.__ReadNode(inc)
                else:
                    return line
        raise StopIteration
    next = __next__  # BBB for python <3
