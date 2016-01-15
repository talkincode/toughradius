

class TestEvent():

    def __init__(self,parent):
        self.parent=parent

    def event_test(self):
        pass


def get_instance(app):
    return TestEvent(app)