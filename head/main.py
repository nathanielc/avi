#!/usr/bin/env python

from direct.showbase.ShowBase import ShowBase
base = ShowBase()

from panda3d.core import NodePath, TextNode, WindowProperties, CompassEffect

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


        self.frame = 0
        self.frames = []
        self.objs = {}
        self.scores = {}

        self.loadFrames(frames_src)
        self.loadMap()

        taskMgr.add(self.updateFrame, "update frame")

        # hide mouse cursor, comment these 3 lines to see the cursor
        #props = WindowProperties()
        #props.setCursorHidden(True)
        #base.win.requestProperties(props)
        #base.disableMouse()  # disable mouse control of the camera



        ## dummy node for camera, we will rotate the dummy node fro camera rotation
        #self.camera_origin = render.attachNewNode('camOrigin')
        #self.camera_origin.reparentTo(render) # inherit transforms
        ##self.camera_origin.setEffect(CompassEffect.make(render)) # NOT inherit rotation

        ## the camera
        #base.camera.reparentTo(self.camera_origin)
        #base.camera.lookAt(self.camera_origin)

        ## camera zooming
        #base.accept('wheel_up', self.wheel_up)
        #base.accept('wheel_down', self.wheel_down)


        ## global vars for camera rotation
        #self.heading = 0
        #self.pitch = 0

        #taskMgr.add(self.cameraTask, 'thirdPersonCameraTask')


    def loadFrames(self, frames_src):
        with gzip.open(frames_src) as f:
            self.frames = head_pb2.Stream()
            self.frames.ParseFromString(f.read())


    def updateFrame(self, task):
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
                elif obj.tex == 1:
                    model.setScale(obj.radius)
                    tex = loader.loadTexture("models/asteroid.jpg")
                elif obj.tex == 2:
                    model.setScale(obj.radius)
                    tex = loader.loadTexture("models/control_point.jpg")
                elif obj.tex == 3:
                    tex = loader.loadTexture("models/bullet.jpg")
                    model.setScale(obj.radius*20)
                else:
                    print "Invalid texture", obj
                    exit(1)
                model.setTexture(tex, 1)
                model.reparentTo(render)
                self.objs[name] = model
            else:
                model = self.objs[name]

            pos = obj.pos
            try:
                model.setPos(pos.x, pos.y, pos.z)
            except:
                pass

        # Remove objects that are no longer in the frame
        toRemove = []
        for obj in self.objs:
            if obj not in alive:
                self.objs[obj].removeNode()
                toRemove.append(obj)

        for obj in toRemove:
            del self.objs[obj]

        # Update Scores
        ypos = -0.1
        for score in frame.score:
            fleet = score.fleet
            if fleet not in self.scores:
                self.scores[fleet] = OnscreenText(
                    text="",
                    parent=base.a2dTopRight,
                    align=TextNode.A_left,
                    style=1,
                    fg=(0.2, 0.8, 0, 1),
                    pos=(-0.5, ypos),
                    scale=.07,
                )

                ypos -=  0.1
            self.scores[fleet].setText("%s\t%d" % (fleet, score.score))

        return task.cont

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
        self.sky.setScale(5000)

    def cameraTask(self, task):

        md = base.win.getPointer(0)

        x = md.getX()
        y = md.getY()

        if base.win.movePointer(0, 300, 300):
            self.heading = self.heading + (x - 300) * 0.5
            self.pitch = self.pitch + (y - 300) * 0.5

        camera.setHpr(self.heading, self.pitch,0)

        return task.cont

    def wheel_up(self):
        base.camera.setPos(base.camera.getPos()+200 * globalClock.getDt())

    def wheel_down(self):
        base.camera.setPos(base.camera.getPos()-200 * globalClock.getDt())





# end class world
w = World(sys.argv[1])
run()
