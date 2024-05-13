import pygame, sys
from pygame.locals import *
import random
import queue


class Property:
    def __init__(self, position, price, value, tax):
        self.position = position
        self.price = price
        self.value = value
        self.tax = tax
        self.owner = None

    def buy(self, player):
        if self.owner:
            return
        if player.charge(self.price):
            self.owner = player
            player.addToInventory(self.position)

    def sell(self):
        if self.owner:
            self.owner.addCash(self.value)
            self.owner.removeFromInventory(self.position)
            self.owner = None

    def charge(self, player):
        if self.owner:
            amount = player.charge(self.tax)
            self.owner.addCash(amount)


class Player:
    def __init__(self, ID):
        self.id = ID
        self.position = 0
        self.cash = 2000
        self.inventory = set()
        self.jailed = False

    def getPosition(self):
        return self.position

    def setPosition(self, pos):
        self.position = pos

    def addCash(self, amount):
        self.cash += amount

    def charge(self, amount):
        if amount > self.cash:
            subtracted = self.cash
            self.cash = 0
            return subtracted

        self.cash -= amount
        return amount

    def addToInventory(self, pos):
        self.inventory.add(pos)

    def removeFromInventory(self, pos):
        self.inventory.discard(pos)

class Game:
    def __init__(self):
        self.queue = queue.Queue()
        self.queue.put(Player(0))
        self.queue.put(Player(1))
        self.current_player = None
        self.properties = {}
        self.initProperties()

    def initProperties(self):
        for pos in [1,2,4,5,6,7,8,9,10,11]:
            price = random.randrange(100, 2000, 100)
            value = price / 2
            tax = price / 4
            self.properties[pos] = Property(pos, price, value, tax)

    def nextPlayer(self):
        player = self.queue.get()
        self.current_player = player
        self.queue.put(player)

    def movePlayer(self):
        amount = random.randint(1,6)
        curr_pos = self.current_player.getPosition()
        next_pos = (curr_pos + amount) % len(self.properties)
        self.current_player.setPosition(next_pos)
        if curr_pos + amount >= len(self.properties):
            self.current_player.addCash(200)

    def handleNewPosition(self):
        pos = self.current_player.getPosition()
        if pos == 3:
            self.current_player.jailed = True

    def handleEvents(self):
        pass


    def start(self):
        while(True):

            self.nextPlayer()
            self.movePlayer()
            self.handleNewPosition()
            if not self.current_player.jailed:
                playing = True
            while(playing):
                self.handleEvents()