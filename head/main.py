#!/usr/bin/env python

# Author: Shao Zhang and Phil Saltzman
# Last Updated: 2015-03-13
#
# This tutorial is intended as a initial panda scripting lesson going over
# display initialization, loading models, placing objects, and the scene graph.
#
# Step 4: In this step, we will load the rest of the planets up to Mars.
# In addition to loading them, we will organize how the planets are grouped
# hierarchically in the scene. This will help us rotate them in the next step
# to give a rough simulation of the solar system.  You can see them move by
# running step_5_complete_solar_system.py.

from direct.showbase.ShowBase import ShowBase
base = ShowBase()

from panda3d.core import NodePath, TextNode
from direct.gui.DirectGui import *
import sys
import json
import time
import object_pb2


class World(object):

    def __init__(self):
        # This is the initialization we had before
        self.title = OnscreenText(  # Create the title
            text="Avi",
            parent=base.a2dBottomRight, align=TextNode.A_right,
            style=1, fg=(1, 1, 1, 1), pos=(-0.1, 0.1), scale=.07)

        base.setBackgroundColor(0, 0, 0)  # Set the background to black
        #base.disableMouse()  # disable mouse control of the camera
        camera.setPos(0, 0, 45)  # Set the camera position (X, Y, Z)
        camera.setHpr(0, -90, 0)  # Set the camera orientation
        #(heading, pitch, roll) in degrees


        self.frame = 0
        self.frames = []
        self.objs = {}

        self.loadFrames()
        self.loadMap()
        self.loop()

    def loop(self):
        count = 0
        rate = 100
        while True:
            if count % rate == 0:
                self.updateEvents()
            count += 1
            taskMgr.step()


    def loadFrames(self):
        with open('frames.dat') as f:
            self.frames = object_pb2.Stream()
            self.frames.ParseFromString(f.read())


    def updateEvents(self):
        print "Frame ", self.frame
        frame = self.frames.frame[self.frame]
        self.frame = (self.frame + 1) % len(self.frames.frame)
        for obj in frame.object:
            name = obj.ID
            model = None
            if name not in self.objs:
                print "New model"
                model = loader.loadModel("models/planet_sphere")
                if obj.tex == 0:
                    model.setScale(50)
                    tex = loader.loadTexture("models/earth_1k_tex.jpg")
                else:
                    tex = loader.loadTexture("models/sun_1k_tex.jpg")
                    model.setScale(5)
                model.setTexture(tex, 1)
                model.reparentTo(render)
                self.objs[name] = model
            else:
                model = self.objs[name]

            pos = obj.pos
            model.setPos(pos.x, pos.y, pos.z)




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

        # These are the same steps we used to load the sun in the last step.
        # Again, we use loader.loadModel since we're using planet_sphere more
        # than once.
        #self.sun = loader.loadModel("models/planet_sphere")
        #self.sun_tex = loader.loadTexture("models/sun_1k_tex.jpg")
        #self.sun.setTexture(self.sun_tex, 1)
        #self.sun.reparentTo(render)
        #self.sun.setScale(3 * self.sizescale)




# end class world

w = World()
