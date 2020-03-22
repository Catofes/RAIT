package rait

import (
	"fmt"
)

func (r *RAIT) SetupLoopback() error {
	helper, err := NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return fmt.Errorf("SetupLoopback: failed to get netns helper: %w", err)
	}
	defer helper.Destroy()
	lo, err := helper.DstHandle.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("SetupLoopback: failed to get loopback interface in specfied netns: %w", err)
	}
	err = helper.DstHandle.LinkSetUp(lo)
	if err != nil {
		return fmt.Errorf("SetupLoopback: failed to bring up loopback interface in specfied netns: %w", err)
	}
	err = helper.DstHandle.AddrAdd(lo, SynthesisAddress(r.Name))
	if err != nil {
		return fmt.Errorf("SetupLoopback: failed to add addr to loopback interface in specfied netns: %w", err)
	}
	return nil
}

func (r *RAIT) DestroyLoopback() error {
	helper, err := NamespaceHelperFromName(r.Namespace)
	if err != nil {
		return fmt.Errorf("DestroyLoopback: failed to get netns helper: %w", err)
	}
	defer helper.Destroy()
	lo, err := helper.DstHandle.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("DestroyLoopback: failed to get loopback interface in specfied netns: %w", err)
	}
	err = helper.DstHandle.LinkSetDown(lo)
	if err != nil {
		return fmt.Errorf("DestroyLoopback: failed to bring down loopback interface in specfied netns: %w", err)
	}
	return nil
}
