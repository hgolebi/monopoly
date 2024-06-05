import pygame, sys
from pygame.locals import *
import random
import queue
import keyboard


class Field:
    def __init__(self, name):
        self.name = name


class Property(Field):
    def __init__(self, name, price, value, tax):
        super().__init__( name)
        self.price = price
        self.value = value
        self.tax = tax
        self.owner = None

    def buy(self, player):
        if self.owner:
            return False
        if player.canAfford(self.price):
            player.charge(self.price)
            self.owner = player
            player.addToInventory(self)
            return True
        return False

    def sell(self):
        if self.owner:
            self.owner.addCash(self.value)
            self.owner.removeFromInventory(self)
            self.owner = None

    def charge(self, player):
        if self.owner:
            amount = player.charge(self.tax)
            self.owner.addCash(amount)


    def getOwner(self):
        return self.owner

class TaxField(Field):
    def __init__(self, tax):
        super().__init__("tax field")
        self.tax = tax

    def getTax(self):
        return self.tax

class GoToJailField(Field):
    def __init__(self):
        super().__init__( "go to jail field")

    def action(self, player):
        player.setJailed(True)

class Jail(Field):
    def __init__(self, bail):
        super().__init__( "jail")
        self.bail = bail


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

    def isJailed(self):
        return self.jailed

    def setJailed(self, is_jailed):
        self.jailed = is_jailed

    def canAfford(self, amount):
        return self.cash >= amount

class Game:
    def __init__(self):
        self.queue = queue.Queue()
        self.queue.put(Player(0))
        self.queue.put(Player(1))
        self.current_player = None
        self.fields = {}
        self.initFields()

    def initFields(self):
        self.fields = (
            Field('start'),
            Property('brown 1', 60, 30, 2),
            Property('brown 2', 60, 30, 4),
            TaxField(200),
            Property('railway 1', 200, 100, 25),
            Property('light blue 1', 100, 50, 6),
            Property('light blue 2', 100, 50, 6),
            Property('light blue 3', 120, 60, 8),
            Jail(50),
            Property('pink 1', 140, 70, 10),
            Property('pink 2', 140, 70, 10),
            Property('pink 3', 160, 80, 12),
            Property("railway 2", 200, 100, 25),
            Property('orange 1', 180, 90, 14),
            Property('orange 2', 180, 90, 14),
            Property('orange 3', 200, 100, 16),
            Field('parking'),
            Property('red 1', 220, 110, 18),
            Property('red 2', 220, 110, 18),
            Property('red 3', 240, 120, 20),
            Property('railway 3', 200, 100, 25),
            Property('yellow 1', 260, 130, 22),
            Property('yellow 2', 260, 130, 22),
            Property('yellow 3', 280, 140, 24),
            GoToJailField(),
            Property('green 1', 300, 150, 26),
            Property('green 2', 300, 150, 26),
            Property('green 3', 320, 260, 28),
            Property('railway', 200, 100, 25),
            Property('dark blue 1', 350, 175, 35),
            TaxField(100),
            Property('dark blue 2', 400, 200, 50),
        )

    def rollDice(self):
        return random.randint(1,6), random.randint(1,6)

    def nextPlayer(self):
        player = self.queue.get()
        self.current_player = player
        self.queue.put(player)

    def movePlayer(self):
        if self.current_player.isJailed():
            return
        amount = random.randint(1,6)
        curr_pos = self.current_player.getPosition()
        next_pos = (curr_pos + amount) % len(self.fields)
        self.current_player.setPosition(next_pos)
        if curr_pos + amount >= len(self.fields):
            self.current_player.addCash(200)

    def handleJailField(self, field):
        if not self.current_player.isJailed():
            return

        while True:
            if keyboard.is_pressed('1'):
                die1, die2 = self.rollDice()
                if die1 == die2:
                    self.current_player.setJailed(False)
                break
            if keyboard.is_pressed('2'):
                if self.current_player.canAfford(field.bail):
                    self.current_player.charge(field.bail)
                    self.current_player.setJailed(False)
                    break



    def handlePropertyField(self, field):
        owner = field.getOwner()
        if not owner:
            while True:
                if keyboard.is_pressed('y'):
                    field.buy()
                    break
                if keyboard.is_pressed('n'):
                    break
        elif owner != self.current_player:
            field.charge(self.current_player)


    def handleFieldAction(self):
        field = self.fields[self.current_player.getPosition()]
        if isinstance(field, GoToJailField):
            self.current_player.setJailed(True)
        elif isinstance(field, Jail):
            self.handleJailField(field)
        elif isinstance(field, TaxField):
            self.current_player.charge(field.getTax)
        elif isinstance(field, Property):
            self.handlePropertyField(field)


    def start(self):
        while(True):
            self.nextPlayer()
            self.movePlayer()
            self.handleFieldAction()
