package ems

/*
#cgo darwin CFLAGS: -I.
#cgo darwin CFLAGS: -I/opt/tibco/ems/ems830/ems/8.3/include/tibems
#cgo darwin LDFLAGS: -L/opt/tibco/ems/ems830/ems/8.3/lib -ltibems64

#include <tibems.h>
tibemsDestination castToDestination(tibemsTemporaryQueue queue) {
  return (tibemsDestination)queue;
}
tibems_bool castToBool(int value) {
	return (tibems_bool)value;
}
tibems_long castToLong(int value) {
  return (tibems_long)value;
}
tibems_int castToInt(int value) {
  return (tibems_int)value;
}

*/
import "C"
import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
	"fmt"
)

type Client interface {
	IsConnected() bool
	Connect() error
	Disconnect() error
	Send(destination string, message string, deliveryDelay int, deliveryMode string, expiration int) error
	SendReceive(destination string, message string, deliveryMode string, expiration int) (string, error)
}

type client struct {
	conn         C.tibemsConnection
	cf           C.tibemsConnectionFactory
	errorContext C.tibemsErrorContext
	status       uint32
	options      ClientOptions
	sync.RWMutex
}

func NewClient(o *ClientOptions) Client {

	c := &client{}
	c.options = *o
	c.status = disconnected

	return c
}

func (c *client) IsConnected() bool {

	c.RLock()
	defer c.RUnlock()

	return c.status == connected

}
func (c *client) Connect() error {

	c.RLock()
	defer c.RUnlock()

	status := C.tibemsErrorContext_Create(&c.errorContext)

	if status != TIBEMS_OK {
		return errors.New("failed to create error context")
	}

	c.cf = C.tibemsConnectionFactory_Create()

	url := c.options.GetServerUrl()

	status = C.tibemsConnectionFactory_SetServerURL(c.cf, C.CString(url.String()))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// create the connection
	status = C.tibemsConnectionFactory_CreateConnection(c.cf, &c.conn, C.CString(c.options.username), C.CString(c.options.password))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// start the connection
	status = C.tibemsConnection_Start(c.conn)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	c.setConnected(connected)

	return nil
}

func (c *client) Disconnect() error {

	c.RLock()
	defer c.RUnlock()

	if c.IsConnected() {

		status := C.tibemsConnection_Stop(c.conn)
		if status != TIBEMS_OK {
			return errors.New("failed to stop connection")
		}

		// close the connection
		status = C.tibemsConnection_Close(c.conn)
		if status != TIBEMS_OK {
			return errors.New("failed to close connection")
		}

		c.setConnected(disconnected)
	}

	return nil
}

func (c *client) SendReceive(destination string, message string, deliveryMode string, expiration int) (string, error) {
	var dest C.tibemsDestination
	var session C.tibemsSession
	var requestor C.tibemsMsgRequestor
	var reqMsg C.tibemsMsg
	var repMsg C.tibemsMsg

	var msg C.tibemsTextMsg

	// create the destination
	status := C.tibemsDestination_Create(&dest, TIBEMS_QUEUE, C.CString(destination))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// create the session
	status = C.tibemsConnection_CreateSession(c.conn, &session, TIBEMS_FALSE, TIBEMS_AUTO_ACKNOWLEDGE)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// create the requestor
	status = C.tibemsMsgRequestor_Create(session, &requestor, dest)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// create the request message
	status = C.tibemsMsg_Create(&reqMsg)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// create the message
	status = C.tibemsTextMsg_Create(&msg)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// set message delivery mode
	var emsDeliveryMode = TIBEMS_NON_PERSISTENT
	if strings.ToLower(deliveryMode) == "persistent" {
		emsDeliveryMode = TIBEMS_PERSISTENT
	} else if strings.ToLower(deliveryMode) == "non_persistent" {
		emsDeliveryMode = TIBEMS_NON_PERSISTENT
	} else if strings.ToLower(deliveryMode) == "reliable" {
		emsDeliveryMode = TIBEMS_RELIABLE
	}

	status = C.tibemsMsg_SetDeliveryMode(msg, C.tibemsDeliveryMode(emsDeliveryMode))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// set message expiration
	status = C.tibemsMsg_SetExpiration(msg, C.castToLong(C.int(expiration)))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// create the reply message
	status = C.tibemsMsg_Create(&repMsg)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// set the message text
	status = C.tibemsTextMsg_SetText(msg, C.CString(message))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// send a request message; wait for a reply
	status = C.tibemsMsgRequestor_Request(requestor, msg, &repMsg)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// Get the string data from the reply text message

	var buf *C.char
	buf = (*C.char)(C.calloc(32768, 1))
	defer C.free(unsafe.Pointer(buf))

	status = C.tibemsTextMsg_GetText(repMsg, &buf)

	replyMessageText := C.GoString(buf)


	fmt.Println("Received JMS Reply Text Message = "+replyMessageText)


	// destroy the request message
	status = C.tibemsMsg_Destroy(reqMsg)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// destroy the requestor
	status = C.tibemsMsgRequestor_Close(requestor)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// destroy the session
	status = C.tibemsSession_Close(session)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}

	// destroy the destination
	status = C.tibemsDestination_Destroy(dest)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return "", errors.New(e)
	}


	return replyMessageText, nil

}

func (c *client) Send(destination string, message string, deliveryDelay int, deliveryMode string, expiration int) error {

	var dest C.tibemsDestination
	var session C.tibemsSession
	var msgProducer C.tibemsMsgProducer
	var txtMsg C.tibemsTextMsg

	// create the destination
	status := C.tibemsDestination_Create(&dest, TIBEMS_QUEUE, C.CString(destination))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// create the session
	status = C.tibemsConnection_CreateSession(c.conn, &session, TIBEMS_FALSE, TIBEMS_AUTO_ACKNOWLEDGE)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// create the producer
	status = C.tibemsSession_CreateProducer(session, &msgProducer, dest)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	status = C.tibemsMsgProducer_SetDeliveryDelay(msgProducer, C.castToLong(C.int(deliveryDelay)))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	var emsDeliveryMode = TIBEMS_NON_PERSISTENT
	if strings.ToLower(deliveryMode) == "persistent" {
		emsDeliveryMode = TIBEMS_PERSISTENT
	} else if strings.ToLower(deliveryMode) == "non_persistent" {
		emsDeliveryMode = TIBEMS_NON_PERSISTENT
	} else if strings.ToLower(deliveryMode) == "reliable" {
		emsDeliveryMode = TIBEMS_RELIABLE
	}

	status = C.tibemsMsgProducer_SetDeliveryMode(msgProducer, C.castToInt(C.int(emsDeliveryMode)))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	status = C.tibemsMsgProducer_SetTimeToLive(msgProducer, C.castToLong(C.int(expiration)))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// create the message
	status = C.tibemsTextMsg_Create(&txtMsg)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// set the message text
	status = C.tibemsTextMsg_SetText(txtMsg, C.CString(message))
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// publish the message
	status = C.tibemsMsgProducer_Send(msgProducer, txtMsg)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// destroy the message
	status = C.tibemsMsg_Destroy(txtMsg)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// destroy the producer
	status = C.tibemsMsgProducer_Close(msgProducer)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// destroy the session
	status = C.tibemsSession_Close(session)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// destroy the destination
	status = C.tibemsDestination_Destroy(dest)
	if status != TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	return nil
}

func (c *client) connectionStatus() uint32 {
	c.RLock()
	defer c.RUnlock()
	status := atomic.LoadUint32(&c.status)
	return status
}

func (c *client) setConnected(status uint32) {
	c.RLock()
	defer c.RUnlock()
	atomic.StoreUint32(&c.status, status)
}

func (c *client) getErrorContext() (string, string) {

	var errorString, stackTrace = "", ""
	var buf *C.char
	defer C.free(unsafe.Pointer(buf))

	C.tibemsErrorContext_GetLastErrorString(c.errorContext, &buf)
	errorString = C.GoString(buf)

	C.tibemsErrorContext_GetLastErrorStackTrace(c.errorContext, &buf)
	stackTrace = C.GoString(buf)

	return errorString, stackTrace

}
