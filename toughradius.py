#!/usr/bin/env python
#coding:utf-8

import sys
import webbrowser
from PyQt4 import QtCore, QtGui, uic

class ToughRadius(QtGui.QMainWindow):
    def __init__(self, *args):
        super(ToughRadius, self).__init__(*args)
        uic.loadUi('tr.ui', self)
        self.codec = QtCore.QTextCodec.codecForName("UTF-8")
        self.rad_stop.setEnabled(False)  
        self.web_stop.setEnabled(False)  

    @QtCore.pyqtSlot()
    def on_web_open_clicked(self):   
         webbrowser.open("http://127.0.0.1:%s"%self.web_port.text())

    @QtCore.pyqtSlot()
    def on_log_clear_clicked(self):   
         self.web_console.clear()
         self.rad_console.clear()       

    @QtCore.pyqtSlot()
    def on_rad_start_clicked(self):
        self.rad_proc =  QtCore.QProcess(self)
        self.rad_proc.setWorkingDirectory("radiusd")
        self.rad_proc.setProcessChannelMode(QtCore.QProcess.MergedChannels);    
        QtCore.QObject.connect(self.rad_proc, QtCore.SIGNAL("finished(int)"), self.rad_finished)
        QtCore.QObject.connect(self.rad_proc, QtCore.SIGNAL("started()"), self.rad_started)
        QtCore.QObject.connect(self.rad_proc, QtCore.SIGNAL("readyReadStandardOutput()"), self.OnRadOutputReady)
        self.rad_proc.start("python",["server.py",
            "-auth",self.auth_port.text() or "1812",
            "-acct",self.acct_port.text() or "1813",
            "-admin",self.admin_port.text() or "1815"])

    @QtCore.pyqtSlot()
    def on_rad_stop_clicked(self):
        if self.rad_proc:
            self.rad_proc.close()


    @QtCore.pyqtSlot()
    def on_web_start_clicked(self):
        self.web_proc =  QtCore.QProcess(self)
        self.web_proc.setWorkingDirectory("console")
        self.web_proc.setProcessChannelMode(QtCore.QProcess.MergedChannels);    
        QtCore.QObject.connect(self.web_proc, QtCore.SIGNAL("finished(int)"), self.web_finished)
        QtCore.QObject.connect(self.web_proc, QtCore.SIGNAL("started()"), self.web_started)
        QtCore.QObject.connect(self.web_proc, QtCore.SIGNAL("readyReadStandardOutput()"), self.OnWebOutputReady)
        self.web_proc.start("python",["admin.py",
            "-http",self.web_port.text() or "1816",
            "-admin",self.admin_port.text() or "1815"])

    @QtCore.pyqtSlot()
    def on_web_stop_clicked(self):
        if self.web_proc:
            self.web_proc.close()  

    @QtCore.pyqtSlot()
    def OnRadOutputReady(self):
        if self.rad_proc and self.show_log.isChecked():   
            self.rad_console.append(self.codec.toUnicode(self.rad_proc.readAllStandardOutput().data()))

    @QtCore.pyqtSlot()
    def OnWebOutputReady(self):
        if self.web_proc and self.show_log.isChecked():   
            self.web_console.append(self.codec.toUnicode(self.web_proc.readAllStandardOutput().data()))            
    
    @QtCore.pyqtSlot()
    def rad_started(self): 
        self.rad_start.setEnabled(False)   
        self.rad_stop.setEnabled(True)   
        self.rad_console.append("\nRadiusServer started\n")
        self.ctabs.setCurrentWidget(self.radtab)

    
    @QtCore.pyqtSlot(int)
    def rad_finished(self, rv):
        self.rad_proc = None   
        self.rad_start.setEnabled(True)   
        self.rad_stop.setEnabled(False)   
        self.rad_console.append("\nRadiusServer stoped\n") 
        self.ctabs.setCurrentWidget(self.radtab)    

    @QtCore.pyqtSlot()
    def web_started(self): 
        self.web_start.setEnabled(False)   
        self.web_stop.setEnabled(True)   
        self.web_console.append("\nWebServer started\n")
        self.ctabs.setCurrentWidget(self.webtab)

    
    @QtCore.pyqtSlot(int)
    def web_finished(self, rv):
        self.web_proc = None   
        self.web_start.setEnabled(True)   
        self.web_stop.setEnabled(False)   
        self.web_console.append("\nWebServer stoped\n")  
        self.ctabs.setCurrentWidget(self.webtab)



app = QtGui.QApplication(sys.argv)
widget = ToughRadius()
widget.show()
sys.exit(app.exec_())