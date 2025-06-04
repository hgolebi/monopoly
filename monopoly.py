import pygame, sys
from pygame.locals import *
import random
import queue
import keyboard


def cash(amount):
    return str(amount) + '$'

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


class Jail(Field):
    def __init__(self, bail):
        super().__init__( "jail")
        self.bail = bail

class Start(Field):
    def __init__(self):
        super().__init__( "start field")

class CarPark(Field):
    def __init__(self):
        super().__init__( "car park")

class Player:
    def __init__(self, ID, cash):
        self.id = ID
        self.position = 0
        self.cash = cash
        self.inventory = set()
        self.jailed = False

    def getPosition(self):
        return self.position

    def setPosition(self, pos):
        self.position = pos

    def addCash(self, amount):
        self.cash += amount

    def charge(self, amount):
        paid = min(self.cash, amount)
        self.cash -= amount
        return paid

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
        self.current_player = Player(0, 1000)
        self.queue = queue.Queue()
        self.queue.put(Player(1, 1000))
        self.fields = {}
        self.initFields()

    def initFields(self):
        self.fields = (
            Start(),
            Property('brown 1', 60, 30, 20),
            Property('brown 2', 60, 30, 40),
            TaxField(200),
            Property('railway 1', 200, 100, 250),
            Property('light blue 1', 100, 50, 60),
            Property('light blue 2', 100, 50, 60),
            Property('light blue 3', 120, 60, 80),
            Jail(50),
            Property('pink 1', 140, 70, 100),
            Property('pink 2', 140, 70, 100),
            Property('pink 3', 160, 80, 120),
            Property("railway 2", 200, 100, 250),
            Property('orange 1', 180, 90, 140),
            Property('orange 2', 180, 90, 140),
            Property('orange 3', 200, 100, 160),
            CarPark(),
            Property('red 1', 220, 110, 180),
            Property('red 2', 220, 110, 180),
            Property('red 3', 240, 120, 200),
            Property('railway 3', 200, 100, 250),
            Property('yellow 1', 260, 130, 220),
            Property('yellow 2', 260, 130, 220),
            Property('yellow 3', 280, 140, 240),
            GoToJailField(),
            Property('green 1', 300, 150, 260),
            Property('green 2', 300, 150, 260),
            Property('green 3', 320, 260, 280),
            Property('railway', 200, 100, 250),
            Property('dark blue 1', 350, 175, 350),
            TaxField(100),
            Property('dark blue 2', 400, 200, 500),
        )

    def rollDice(self):
        die1, die2 = random.randint(1,6), random.randint(1,6)
        print('You rolled ', die1, ' and ', die2)
        return die1, die2

    def nextPlayer(self):
        if self.current_player.cash >= 0:
            self.queue.put(self.current_player)
        else:
            print("Player", self.current_player.id, 'has bankrupted..')
        player = self.queue.get()
        self.current_player = player
        print("")
        print('Now playing: player ', self.current_player.id)
        print('Available cash: ', cash(self.current_player.cash))

    def movePlayer(self):
        if self.current_player.isJailed():
            print("You're jailed!")
            return
        print("Press SPACEBAR to roll dice!")
        keyboard.wait('space')
        die1, die2 = self.rollDice()

        amount = die1 + die2
        curr_pos = self.current_player.getPosition()
        next_pos = (curr_pos + amount) % len(self.fields)
        self.current_player.setPosition(next_pos)
        print('You moved to ', self.fields[next_pos].name)
        if curr_pos + amount >= len(self.fields):
            print("You get free 200$ for crossing the Start!")
            self.current_player.addCash(200)

    def handleJailField(self, field):
        if not self.current_player.isJailed():
            print("You're just visiting.")
            return

        print("To get out of the jail you have to pay 50$ or roll the double.")
        print("Press 1 to roll dice")
        print("Press 2 to pay 50$")
        while True:
            if keyboard.is_pressed('1'):
                die1, die2 = self.rollDice()
                if die1 == die2:
                    print("You're no longer jailed!")
                    self.current_player.setJailed(False)
                else:
                    print("You failed to get out of jail!")
                break
            if keyboard.is_pressed('2'):
                if self.current_player.canAfford(field.bail):
                    self.current_player.charge(field.bail)
                    self.current_player.setJailed(False)
                    print("You're no longer jailed!")
                    break
                print('You cannot afford it!')


    def handlePropertyField(self, field):
        owner = field.getOwner()
        if not owner:
            print('Do you want to buy this property for', cash(field.price) + '? (Y/n)')
            while True:
                if keyboard.is_pressed('y'):
                    if field.buy(self.current_player):
                        print("Property bought")
                    else:
                        print("You cannot afford it!")
                    break
                if keyboard.is_pressed('n'):
                    break
        elif owner != self.current_player:
            print('This property is owned by player', owner.id )
            print("You have to pay him", cash(field.tax))
            field.charge(self.current_player)
        else:
            print("You are the owner!")


    def handleFieldAction(self):
        field = self.fields[self.current_player.getPosition()]
        if isinstance(field, GoToJailField):
            print("Unfortunately you got jailed..")
            self.current_player.setJailed(True)
            self.current_player.setPosition(8)
        elif isinstance(field, Jail):
            self.handleJailField(field)
        elif isinstance(field, TaxField):
            print("Unfortunately you have to pay", cash(field.getTax()), 'tax')
            self.current_player.charge(field.getTax())
        elif isinstance(field, Property):
            self.handlePropertyField(field)
        elif isinstance(field, CarPark):
            print("You can park here for free!")

    def start(self):
        print("\n\n\n")
        print("Welcome to Monopoly!")
        while(True):
            self.nextPlayer()
            if self.queue.empty():
                print("Player", self.current_player.id, 'wins the game! Congratulations!')
                return
            self.movePlayer()
            self.handleFieldAction()




if __name__ == "__main__":
    game = Game()
    game.start()