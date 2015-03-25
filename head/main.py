#!/usr/bin/env python

from direct.showbase.ShowBase import ShowBase
base = ShowBase()

from panda3d.core import NodePath, TextNode
from direct.gui.DirectGui import *
import sys
import json
import time
import gzip

import head_pb2


class World(object):

    def __init__(self, frames_src):
        # This is the initialization we had before
        self.title = OnscreenText(  # Create the title
            text="Avi",
            parent=base.a2dBottomRight, align=TextNode.A_right,
            style=1, fg=(1, 1, 1, 1), pos=(-0.1, 0.1), scale=.07)

        base.setBackgroundColor(0, 0, 0)  # Set the background to black
        #base.disableMouse()  # disable mouse control of the camera
        camera.setPos(150, 0, 0)  # Set the camera position (X, Y, Z)
        camera.setHpr(0, -90, 0)  # Set the camera orientation
        #(heading, pitch, roll) in degrees


        self.frame = 0
        self.frames = []
        self.objs = {}

        self.loadFrames(frames_src)
        self.loadMap()
        self.loop()

    def loop(self):
        count = 0
        rate = 2
        while True:
            if count % rate == 0:
                self.updateEvents()
            count += 1
            taskMgr.step()


    def loadFrames(self, frames_src):
        with gzip.open(frames_src) as f:
            self.frames = head_pb2.Stream()
            self.frames.ParseFromString(f.read())


    def updateEvents(self):
        print "Frame ", self.frame
        frame = self.frames.frame[self.frame]
        self.frame = (self.frame + 1) % len(self.frames.frame)
        alive = {}
        for obj in frame.object:
            name = obj.ID
            alive[name] = True
            model = None
            if name not in self.objs:
                model = loader.loadModel("models/sphere")
                if obj.tex == 0:
                    model.setScale(obj.radius)
                    tex = loader.loadTexture("models/%s.jpg" % obj.tex_custom)
                else:
                    tex = loader.loadTexture("models/steel.jpg")
                    model.setScale(obj.radius*10)
                model.setTexture(tex, 1)
                model.reparentTo(render)
                self.objs[name] = model
            else:
                model = self.objs[name]

            pos = obj.pos
            model.setPos(pos.x, pos.y, pos.z)

        # Remove objects that are no longer in the frame
        toRemove = []
        for obj in self.objs:
            if obj not in alive:
                self.objs[obj].removeNode()
                toRemove.append(obj)

        for obj in toRemove:
            del self.objs[obj]

    def loadMap(self):
        # These are the same steps used to load the sky model that we used in the
        # last step
        # Load the model for the sky
        self.sky = loader.loadModel("models/solar_sky_sphere")
        # Load the texture for the sky.
        self.sky_tex = loader.loadTexture("models/stars_1k_tex.jpg")
        # Set the sky texture to the sky model
        self.sky.setTexture(self.sky_tex, 1)
        # Parent the sky model to the render node so that the sky is rendered
        self.sky.reparentTo(render)
        # Scale the size of the sky.
        self.sky.setScale(1000)

# end class world
w = World(sys.argv[1])
