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
	"github.com/mmussett/jmsproxy/tibems"
	"sync"
	"sync/atomic"
	"unsafe"
	"strings"
)

type TibemsErrorContext struct {
	errorContext C.tibemsErrorContext
}

type TibemsConnectionFactory struct {
	factory C.tibemsConnectionFactory
}

type TibemsConnection struct {
	connection C.tibemsConnection
}

type Queue struct {
	destination C.tibemsQueue
}

type Topic struct {
	destination C.tibemsTopic
}

type Session struct {
	session C.tibemsSession
}

type QueueSession struct {
	session C.tibemsQueueSession
}

type TopicSession struct {
	session C.tibemsTopicSession
}

type MessageProducer struct {
	messageProducer C.tibemsMsgProducer
}

type Destination struct {
	destination C.tibemsDestination
}

type TextMsg struct {
	msg C.tibemsTextMsg
}

type Msg struct {
	msg C.tibemsMsg
}

type MessageRequestor struct {
	requestor C.tibemsMsgRequestor
}

type TemporaryQueue struct {
	tempQueue C.tibemsTemporaryQueue
}

type QueueReceiver struct {
	queueReceiver C.tibemsQueueReceiver
}

type Boolean struct {
	boolean C.tibems_bool
}

type Client interface {
	IsConnected() bool
	Connect() error
	Disconnect() error
	Send(destination string, message string, deliveryDelay int, deliveryMode string, expiration int) error
}

type client struct {
	conn         TibemsConnection
	cf           TibemsConnectionFactory
	errorContext TibemsErrorContext
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

	status := C.tibemsErrorContext_Create(&c.errorContext.errorContext)

	if status != TIBEMS_OK {
		return errors.New("failed to create error context")
	}

	c.cf.factory = C.tibemsConnectionFactory_Create()

	url := c.options.GetServerUrl()

	status = C.tibemsConnectionFactory_SetServerURL(c.cf.factory, C.CString(url.String()))
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// create the connection
	status = C.tibemsConnectionFactory_CreateConnection(c.cf.factory, &c.conn.connection, C.CString(c.options.username), C.CString(c.options.password))
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// start the connection
	status = C.tibemsConnection_Start(c.conn.connection)
	if status != tibems.TIBEMS_OK {
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

		status := C.tibemsConnection_Stop(c.conn.connection)
		if status != TIBEMS_OK {
			return errors.New("failed to stop connection")
		}

		// close the connection
		status = C.tibemsConnection_Close(c.conn.connection)
		if status != tibems.TIBEMS_OK {
			return errors.New("failed to close connection")
		}

		c.setConnected(disconnected)
	}

	return nil
}

func (c *client) Send(destination string, message string, deliveryDelay int, deliveryMode string, expiration int) error {

	dest := new(Destination)
	session := new(Session)
	msgProducer := new(MessageProducer)
	txtMsg := new(TextMsg)

	// create the destination
	status := C.tibemsDestination_Create(&dest.destination, tibems.TIBEMS_QUEUE, C.CString(destination))
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// create the session
	status = C.tibemsConnection_CreateSession(c.conn.connection, &session.session, tibems.TIBEMS_FALSE, tibems.TIBEMS_AUTO_ACKNOWLEDGE)
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// create the producer
	status = C.tibemsSession_CreateProducer(session.session, &msgProducer.messageProducer, dest.destination)
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	status = C.tibemsMsgProducer_SetDeliveryDelay(msgProducer.messageProducer, C.castToLong(C.int(deliveryDelay)))
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	var emsDeliveryMode = TIBEMS_NON_PERSISTENT
	if strings.ToLower(deliveryMode)=="persistent"  {
		emsDeliveryMode = TIBEMS_PERSISTENT
	} else if strings.ToLower(deliveryMode)=="non_persistent" {
		emsDeliveryMode = TIBEMS_NON_PERSISTENT
	}

	status = C.tibemsMsgProducer_SetDeliveryMode(msgProducer.messageProducer, C.castToInt(C.int(emsDeliveryMode)))
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	status = C.tibemsMsgProducer_SetTimeToLive(msgProducer.messageProducer, C.castToLong(C.int(expiration)))
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// create the message
	status = C.tibemsTextMsg_Create(&txtMsg.msg)
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// set the message text
	status = C.tibemsTextMsg_SetText(txtMsg.msg, C.CString(message))
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// publish the message
	status = C.tibemsMsgProducer_Send(msgProducer.messageProducer, txtMsg.msg)
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// destroy the message
	status = C.tibemsMsg_Destroy(txtMsg.msg)
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// destroy the producer
	status = C.tibemsMsgProducer_Close(msgProducer.messageProducer)
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// destroy the session
	status = C.tibemsSession_Close(session.session)
	if status != tibems.TIBEMS_OK {
		e, _ := c.getErrorContext()
		return errors.New(e)
	}

	// destroy the destination
	status = C.tibemsDestination_Destroy(dest.destination)
	if status != tibems.TIBEMS_OK {
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

	C.tibemsErrorContext_GetLastErrorString(c.errorContext.errorContext, &buf)
	errorString = C.GoString(buf)

	C.tibemsErrorContext_GetLastErrorStackTrace(c.errorContext.errorContext, &buf)
	stackTrace = C.GoString(buf)

	return errorString, stackTrace

}
