package main

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/spiffe/spire/pkg/agent/plugin/workloadattestor"
	"github.com/spiffe/spire/pkg/common/catalog"
	"github.com/spiffe/spire/proto/spire/common"
	spi "github.com/spiffe/spire/proto/spire/common/plugin"
)

const pluginName = "systemd"

func BuiltIn() catalog.Plugin {
	return builtin(New())
}

func builtin(p *Plugin) catalog.Plugin {
	return catalog.MakePlugin(pluginName, workloadattestor.PluginServer(p))
}

type Plugin struct {
	workloadattestor.UnsafeWorkloadAttestorServer
}

func New() *Plugin {
	//New method should return the Plugin Type being implemented
	return &Plugin{}
}

func (p *Plugin) Configure(ctx context.Context, req *spi.ConfigureRequest) (*spi.ConfigureResponse, error) {
	return &spi.ConfigureResponse{}, nil
}

func (p *Plugin) GetPluginInfo(ctx context.Context, req *spi.GetPluginInfoRequest) (*spi.GetPluginInfoResponse, error) {
	return &spi.GetPluginInfoResponse{}, nil
}

func (p *Plugin) Attest(ctx context.Context, req *workloadattestor.AttestRequest) (*workloadattestor.AttestResponse, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	// Get the unit for the given PID from the systemd service.
	call := conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1").CallWithContext(ctx, "org.freedesktop.systemd1.Manager.GetUnitByPID", 0, uint(req.Pid))
	var unitPath dbus.ObjectPath
	if err := call.Store(&unitPath); err != nil {
		return nil, err
	}

	var selectors []*common.Selector

	// Get the location of the service file
	fragmentPathVariant, err := conn.Object("org.freedesktop.systemd1", unitPath).GetProperty("org.freedesktop.systemd1.Unit.FragmentPath")
	if err != nil {
		return nil, err
	}
	fragmentPath, ok := fragmentPathVariant.Value().(string)
	if !ok {
		return nil, fmt.Errorf("Returned fragment path was not a string: %v", fragmentPathVariant.String())
	}
	selectors = append(selectors, makeSelector("fragmentPath", fragmentPath))

	// TODO(zecke): Add other interesting bits of the unit.

	return &workloadattestor.AttestResponse{
		Selectors: selectors,
	}, nil
}

func makeSelector(kind, value string) *common.Selector {
	return &common.Selector{
		Type:  pluginName,
		Value: fmt.Sprintf("%s:%s", kind, value),
	}
}

func main() {
	catalog.PluginMain(BuiltIn())
}
