/* 
 * Copyright (c) 2001-2009 TIBCO Software Inc. 
 * All rights reserved.
 * For more information, please contact:
 * TIBCO Software Inc., Palo Alto, California, USA
 * 
 * $Id: tibems.h 41869 2009-08-28 00:06:10Z $
 * 
 */

#ifndef _INCLUDED_tibems_h
#define _INCLUDED_tibems_h

#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#ifdef __VMS
#include "emsvms.h"
#endif

#include "types.h"
#include "status.h"
#include "dest.h"
#include "confact.h"
#include "conn.h"
#include "tconn.h"
#include "qconn.h"
#include "conmeta.h"
#include "tsess.h"
#include "qsess.h"
#include "msgreq.h"
#include "msgprod.h"
#include "pub.h"
#include "send.h"
#include "msgcons.h"
#include "sub.h"
#include "recv.h"
#include "msg.h"
#include "tmsg.h"
#include "bmsg.h"
#include "smsg.h"
#include "omsg.h"
#include "mmsg.h"
#include "qbrow.h"
#include "xares.h"
#include "cssl.h"
#include "clookup.h"
#include "emserr.h"

#if defined(__cplusplus)
extern "C" {
#endif

/*
 * Multicast exception listener callback
 */
typedef void
(*tibemsMulticastExceptionCallback) (
    tibemsConnection    connection,
    tibemsSession       session,
    tibemsMsgConsumer   consumer,
    tibems_status       status,
    const char*         description,
    void*               closure);

extern const char*
tibems_Version(void);

extern tibems_int
tibems_GetSocketReceiveBufferSize(void);

extern tibems_status
tibems_SetSocketReceiveBufferSize(
    tibems_int  kilobytes);

extern tibems_int
tibems_GetSocketSendBufferSize(void);

extern tibems_status
tibems_SetSocketSendBufferSize(
    tibems_int  kilobytes);

extern tibems_int
tibems_GetConnectAttemptCount(void);

extern tibems_status
tibems_SetConnectAttemptCount(
    tibems_int  count);

extern tibems_int
tibems_GetConnectAttemptDelay(void);

extern tibems_status
tibems_SetConnectAttemptDelay(
    tibems_int  delay);

extern tibems_int
tibems_GetReconnectAttemptCount(void);

extern tibems_status
tibems_SetReconnectAttemptCount(
    tibems_int  count);

extern tibems_int
tibems_GetReconnectAttemptDelay(void);

extern tibems_status
tibems_SetReconnectAttemptDelay(
    tibems_int  delay);

extern tibems_status
tibems_SetConnectAttemptTimeout(
    tibems_int timeout);

extern tibems_int
tibems_GetConnectAttemptTimeout(void);

extern tibems_status
tibems_SetReconnectAttemptTimeout(
    tibems_int timeout);

extern tibems_int
tibems_GetReconnectAttemptTimeout(void);

extern tibems_status
tibems_SetMulticastEnabled(
    tibems_bool enabled);

extern tibems_bool
tibems_GetMulticastEnabled(void);

extern tibems_status
tibems_SetMulticastDaemon(
    const char* port);

extern const char*
tibems_GetMulticastDaemon(void);

extern tibems_status
tibems_setExceptionOnFTSwitch(
    tibems_bool callExceptionListener);

extern tibems_status
tibems_getExceptionOnFTSwitch(
    tibems_bool *isSet);

extern tibems_status
tibems_SetExceptionOnFTEvents(
    tibems_bool callExceptionListener);

extern tibems_status
tibems_GetExceptionOnFTEvents(
    tibems_bool *isSet);

extern tibems_status
tibems_SetTraceFile(
    const char* fileName);

extern tibems_status
tibems_IsConsumerMulticast(
    tibemsMsgConsumer consumer,
    tibems_bool *isMulticast);

extern tibems_status
tibems_SetMulticastExceptionListener(
    tibemsMulticastExceptionCallback listener,
    void*                            closure);

extern tibems_status
tibems_SetAllowCloseInCallback(
    tibems_bool allow);

extern tibems_status
tibems_GetAllowCloseInCallback(
    tibems_bool *isAllowed);

#ifdef  __cplusplus
}
#endif

#endif /* _INCLUDED_tibems_h */
