事件类型（EVENT CODE对应的含义）

“0 BOOTSTAP” 指出由于CPE第一次安装或是ACS的URL改变而引起会话建立。这种特殊的情况有：

1）出厂后CWMP端第一次与ACS连接；
2）出厂设置后，CWMP端第一次与ACS连接；
3）由于某种原因ACS的URL改变后CWMP端第一次与ACS连接。
注意，BOOTSTARP可能和其他事件代码一起组成是事件代码组，例如，在出厂后CPE初始启动时，CPE发送BOOTSTARP和BOOT事件代码。

“1 BOOT” 当给电或是复位时引起的会话建立，这包括初始系统启动或是由于其他原因的再启动，包括用Reboot方法，但是不是从待机状态醒来。

“2 PERIODIC” 在周期通知间隔时会话建立

“3 SCHEDULED” 由于调用ScheduleInform 方法会话建立，这种事件必须只能用“M ScheduleInform”。

“4 VALUE CHANGE” 指出从上次成功的Inform后，具有Passive和Active通知属性的一个或多个参数的值发生了改变，如果这个事件代码在事件组中，所有修改的参数必须被包含在Inform的参数列表中，如果这个事件被丢弃，那么这些修改的参数也应该同时被丢弃。

“5 KICKED” 指出会话建立的目的是网页验证，并且Kicked 方法会在这个会话中调用一次或多次。

“6 CONNECTION REQUEST” 由于ACS 发送了连接请求而使会话建立。

“7 TRANSFER COMPLETE” 由于先前请求的下载或上传完成而引起会话建立，TransferComplete方法会在这个会话中调用一次或多次。这个事件代码必须用“M Download”，“M ScheduleDownload”，或者是"M Upload" etc。

“8 DIAGNOSTICS COMPLETE” 当完成了一个或多个由ACS启动的诊断，CPE会用该事件码重新建立起一个连接。

“9 REQUEST DOWNLOAD” 为了调用RequestDownload方法二发起的会话。

“10 AUTONOMOUS TRANSFER COMPLETE” 当不是由ACS请求的上传或下载完成而引起的会话建立（成功或是不成功），Autonmous TransferComplete 方法会在这个会话中调用一次或多次。

“11 DU STATE CHANGE COMPLETE”, 为了表明先前请求的DU state改变完成而建立的会话，不管成功与否，DUStateChangeComplete方法会在这个会话中调用。这个方法必须用“M ChangeDUState”

“12 AUTONMOUS DU STATE CHANGE COMPLETE” 会话建立是要通知ACS DU state改变完成了， 而这个改变不是由于调用ChangeDUState 方法的请求，DUStateChangeComplete方法会在这个会话中调用。

“13 WAKE UP” 由于CPE从待机中苏醒而建立的会话。

“M Reboot” 由于ACS调用了Reboot RPC，而促使CPE 重新启动，重叠的事件会引起“1 BOOT”事件代码。

“M ScheduleInform” ACS 请求了一个安排通知。

“M Download ” ACS 请求下载

“M ScheduleDownload” ACS请求计划下载

“M Upload” ACS 请求上传

”M ChangeDUState” ACS用ChangeDUState方法请求DU 状态改变。

“M” <vendor specific method>

“X”<VENDOR>" "<event>

