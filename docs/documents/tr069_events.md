# Event type (meaning corresponding to EVENT CODE)

"0 BOOTSTAP" indicates that the session was established due to the first installation of the CPE or a change in the ACS URL. Such special situations are:

1) The CWMP terminal is connected to the ACS for the first time after leaving the factory;
2) After factory setting, the CWMP terminal connects to ACS for the first time;
3) For some reason, the CWMP side connects to ACS for the first time after the ACS URL changes.
Note that BOOTSTARP may form an event code group together with other event codes. For example, when the CPE is initially started after leaving the factory, the CPE sends BOOTSTARP and BOOT event codes.

"1 BOOT" Session establishment caused by power-on or reset. This includes initial system startup or restart due to other reasons, including using the Reboot method, but not waking up from standby.

"2 PERIODIC" session established at periodic notification interval

"3 SCHEDULED" Since the session is established by calling the ScheduleInform method, this event must only be used with "M ScheduleInform".

"4 VALUE CHANGE" indicates that the value of one or more parameters with Passive and Active notification properties has changed since the last successful Inform. If this event code is in the event group, all modified parameters must be included in the Inform In the parameter list, if this event is discarded, then these modified parameters should also be discarded at the same time.

"5 KICKED" indicates that the purpose of session establishment is web page verification, and the Kicked method will be called one or more times in this session.

"6 CONNECTION REQUEST" The session is established due to a connection request sent by ACS.

"7 TRANSFER COMPLETE" The TransferComplete method will be called one or more times in this session due to the completion of a previously requested download or upload. This event code must use "M Download", "M ScheduleDownload", or "M Upload" etc.

"8 DIAGNOSTICS COMPLETE" When one or more diagnostics initiated by the ACS are completed, the CPE will use this event code to re-establish a connection.

"9 REQUEST DOWNLOAD" is a session initiated by calling RequestDownload method two.

"10 AUTONOMOUS TRANSFER COMPLETE" When a session is established (successfully or unsuccessfully) that is not caused by the completion of an upload or download requested by ACS, the Autonmous TransferComplete method will be called one or more times in this session.

"11 DU STATE CHANGE COMPLETE", a session established to indicate the completion of the previously requested DU state change. Regardless of success or failure, the DUStateChangeComplete method will be called in this session. This method must use "M ChangeDUState"

"12 AUTONMOUS DU STATE CHANGE COMPLETE" Session establishment is to notify ACS that the DU state change is completed, and this change is not due to a request to call the ChangeDUState method. The DUStateChangeComplete method will be called in this session.

"13 WAKE UP" Session established due to CPE waking up from standby.

"M Reboot" Since the ACS calls the Reboot RPC, prompting the CPE to restart, the overlapping event will cause a "1 BOOT" event code.

"M ScheduleInform" ACS requested a schedule notification.

“M Download” ACS request download

"M ScheduleDownload" ACS requests schedule download

“M Upload” ACS request upload

"M ChangeDUState" ACS uses the ChangeDUState method to request a DU state change.

"M" 

"X"" "