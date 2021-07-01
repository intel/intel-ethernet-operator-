/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/otcshare/intel-ethernet-operator/pkg/rpc/v1/flow"
)

func validateRteFlowItemEth(itemName string, spec, item *any.Any) error {
	specItem := new(flow.RteFlowItemEth)
	if spec != nil {
		if err := ptypes.UnmarshalAny(spec, specItem); err != nil {
			return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
		}
	}

	ethItem := new(flow.RteFlowItemEth)
	if err := ptypes.UnmarshalAny(item, ethItem); err != nil {
		return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
	}
	var specDstAddrInt, specSrcAddrInt uint32 = 0, 0
	var err error
	if specItem != nil {
		if ethItem.Type != 0 {
			if err = validateUint32ItemField(itemName, "type", specItem.Type, ethItem.Type); err != nil {
				return fmt.Errorf("validateUint32ItemField() error: %v", err)
			}
		}
		if ethItem.Dst != nil {
			if specItem.Dst != nil {
				specDstAddrInt, err = encodeHexMacAddress(specItem.Dst.AddrBytes)
				if err != nil {
					return fmt.Errorf("invalid spec: %v, %v", specItem.Dst.AddrBytes, err)
				}
			}
			dstAddrInt, err := encodeHexMacAddress(ethItem.Dst.AddrBytes)
			if err != nil {
				return fmt.Errorf("invalid %s: %v, %v", itemName, ethItem.Dst.AddrBytes, err)
			}
			if err := validateUint32ItemField(itemName, "dst", specDstAddrInt, dstAddrInt); err != nil {
				return err
			}
		}
		if ethItem.Src != nil {
			if specItem.Src != nil {
				specSrcAddrInt, err = encodeHexMacAddress(specItem.Src.AddrBytes)
				if err != nil {
					return fmt.Errorf("invalid spec: %v, %v", specItem.Src.AddrBytes, err)
				}
			}
			srcAddrInt, err := encodeHexMacAddress(ethItem.Src.AddrBytes)
			if err != nil {
				return fmt.Errorf("invalid %s: %v, %v", itemName, ethItem.Src.AddrBytes, err)
			}
			if err := validateUint32ItemField(itemName, "src", specSrcAddrInt, srcAddrInt); err != nil {
				return err
			}
		}
	}
	if ethItem != nil {
		if ethItem.Type > 0xFFFF {
			return fmt.Errorf("invalid type (%d), must be equal or lower than 65,535", ethItem.Type)
		}
	}

	return nil
}

func encodeHexMacAddress(addr []byte) (uint32, error) {
	intAddr := make([]byte, hex.EncodedLen(len(addr)))
	hex.Encode(intAddr, addr)
	parseAddr, err := strconv.ParseInt(string(intAddr[:]), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing mac address: %v", err)
	}
	return uint32(parseAddr), nil
}

func validateRteFlowItemVlan(itemName string, spec, item *any.Any) error {
	specItem := new(flow.RteFlowItemVlan)
	if spec != nil {
		if err := ptypes.UnmarshalAny(spec, specItem); err != nil {
			return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
		}
	}

	vlanItem := new(flow.RteFlowItemVlan)
	if err := ptypes.UnmarshalAny(item, vlanItem); err != nil {
		return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
	}

	if specItem != nil {
		if err := validateUint32ItemField(itemName, "tci", specItem.Tci, vlanItem.Tci); err != nil {
			return err
		}
		if err := validateUint32ItemField(itemName, "inner_type", specItem.InnerType, vlanItem.InnerType); err != nil {
			return err
		}
	}
	if vlanItem != nil {
		if vlanItem.Tci > 0xFFFF {
			return fmt.Errorf("invalid tci (%d), must be equal or lower than 65,535", vlanItem.Tci)
		}
		if vlanItem.InnerType > 0xFFFF {
			return fmt.Errorf("invalid inner_type (%d), must be equal or lower than 65,535", vlanItem.InnerType)
		}
	}
	return nil
}

func validateRteFlowItemIpv4(itemName string, spec, item *any.Any) error {
	specItem := new(flow.RteFlowItemIpv4)
	if spec != nil {
		if err := ptypes.UnmarshalAny(spec, specItem); err != nil {
			return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
		}
	}

	ipv4Item := new(flow.RteFlowItemIpv4)
	if err := ptypes.UnmarshalAny(item, ipv4Item); err != nil {
		return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
	}

	if specItem != nil {
		if specItem.Hdr != nil {
			if err := validateUint32ItemField(itemName, "version_ihl", specItem.Hdr.VersionIhl, ipv4Item.Hdr.VersionIhl); err != nil {
				return err
			}
			if err := validateUint32ItemField(itemName, "type_of_service", specItem.Hdr.TypeOfService, ipv4Item.Hdr.TypeOfService); err != nil {
				return err
			}
			if err := validateUint32ItemField(itemName, "total_length", specItem.Hdr.TotalLength, ipv4Item.Hdr.TotalLength); err != nil {
				return err
			}
			if err := validateUint32ItemField(itemName, "packet_id", specItem.Hdr.PacketId, ipv4Item.Hdr.PacketId); err != nil {
				return err
			}
			if err := validateUint32ItemField(itemName, "fragment_offset", specItem.Hdr.FragmentOffset, ipv4Item.Hdr.FragmentOffset); err != nil {
				return err
			}
			if err := validateUint32ItemField(itemName, "time_to_live", specItem.Hdr.TimeToLive, ipv4Item.Hdr.TimeToLive); err != nil {
				return err
			}
			if err := validateUint32ItemField(itemName, "next_proto_id", specItem.Hdr.NextProtoId, ipv4Item.Hdr.NextProtoId); err != nil {
				return err
			}
			if err := validateUint32ItemField(itemName, "hdr_checksum", specItem.Hdr.HdrChecksum, ipv4Item.Hdr.HdrChecksum); err != nil {
				return err
			}
			if err := validateUint32ItemField(itemName, "src_addr", specItem.Hdr.SrcAddr, ipv4Item.Hdr.SrcAddr); err != nil {
				return err
			}
			if err := validateUint32ItemField(itemName, "dst_addr", specItem.Hdr.DstAddr, ipv4Item.Hdr.DstAddr); err != nil {
				return err
			}
		}
	}
	if ipv4Item.Hdr != nil {
		if ipv4Item.Hdr.VersionIhl > 0xFF {
			return fmt.Errorf("invalid version_ihl (%d), must be equal or lower than 255", ipv4Item.Hdr.VersionIhl)
		}
		if ipv4Item.Hdr.TypeOfService > 0xFF {
			return fmt.Errorf("invalid type_of_service (%d), must be equal or lower than 255", ipv4Item.Hdr.TypeOfService)
		}
		if ipv4Item.Hdr.TotalLength > 0xFFFF {
			return fmt.Errorf("invalid total_length (%d), must be equal or lower than 65,535", ipv4Item.Hdr.TotalLength)
		}
		if ipv4Item.Hdr.PacketId > 0xFFFF {
			return fmt.Errorf("invalid packet_id (%d), must be equal or lower than 65,535", ipv4Item.Hdr.PacketId)
		}
		if ipv4Item.Hdr.FragmentOffset > 0xFFFF {
			return fmt.Errorf("invalid fragment_offset (%d), must be equal or lower than 65,535", ipv4Item.Hdr.FragmentOffset)
		}
		if ipv4Item.Hdr.TimeToLive > 0xFF {
			return fmt.Errorf("invalid time_to_live (%d), must be equal or lower than 255", ipv4Item.Hdr.TimeToLive)
		}
		if ipv4Item.Hdr.NextProtoId > 0xFF {
			return fmt.Errorf("invalid next_proto_id (%d), must be equal or lower than 255", ipv4Item.Hdr.NextProtoId)
		}
		if ipv4Item.Hdr.HdrChecksum > 0xFF {
			return fmt.Errorf("invalid hdr_checksum (%d), must be equal or lower than 255", ipv4Item.Hdr.HdrChecksum)
		}
	}
	return nil
}

func validateRteFlowItemUdp(itemName string, spec, item *any.Any) error {
	specItem := new(flow.RteFlowItemUdp)
	if spec != nil {
		if err := ptypes.UnmarshalAny(spec, specItem); err != nil {
			return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
		}
	}

	udpItem := new(flow.RteFlowItemUdp)
	if err := ptypes.UnmarshalAny(item, udpItem); err != nil {
		return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
	}

	if specItem != nil {
		if specItem.Hdr != nil {
			if udpItem.Hdr.SrcPort != 0 {
				if err := validateUint32ItemField(itemName, "src_port", specItem.Hdr.SrcPort, udpItem.Hdr.SrcPort); err != nil {
					return err
				}
			}
			if udpItem.Hdr.DstPort != 0 {
				if err := validateUint32ItemField(itemName, "dst_port", specItem.Hdr.DstPort, udpItem.Hdr.DstPort); err != nil {
					return err
				}
			}
			if udpItem.Hdr.DgramLen != 0 {
				if err := validateUint32ItemField(itemName, "dgram_len", specItem.Hdr.DgramLen, udpItem.Hdr.DgramLen); err != nil {
					return err
				}
			}
			if udpItem.Hdr.DgramCksum != 0 {
				if err := validateUint32ItemField(itemName, "dgram_cksum", specItem.Hdr.DgramCksum, udpItem.Hdr.DgramCksum); err != nil {
					return err
				}
			}
		}
	}
	if udpItem.Hdr != nil {
		if udpItem.Hdr.SrcPort > 0xFFFF {
			return fmt.Errorf("invalid src_port (%d), must be equal or lower than 65,535", udpItem.Hdr.SrcPort)
		}
		if udpItem.Hdr.DstPort > 0xFFFF {
			return fmt.Errorf("invalid dst_port (%d), must be equal or lower than 65,535", udpItem.Hdr.DstPort)
		}
		if udpItem.Hdr.DgramLen > 0xFFFF {
			return fmt.Errorf("invalid dgram_len (%d), must be equal or lower than 65,535", udpItem.Hdr.DgramLen)
		}
		if udpItem.Hdr.DgramCksum > 0xFFFF {
			return fmt.Errorf("invalid dgram_cksum (%d), must be equal or lower than 65,535", udpItem.Hdr.DgramCksum)
		}
	}

	return nil
}

func validateRteFlowItemPppoe(itemName string, spec, item *any.Any) error {
	specItem := new(flow.RteFlowItemPppoe)
	if spec != nil {
		if err := ptypes.UnmarshalAny(spec, specItem); err != nil {
			return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
		}
	}

	pppoeItem := new(flow.RteFlowItemPppoe)
	if err := ptypes.UnmarshalAny(item, pppoeItem); err != nil {
		return fmt.Errorf("could not unmarshal %s: %s", itemName, err)
	}

	if specItem != nil {
		if specItem.VersionType != 0 {
			if err := validateUint32ItemField(itemName, "version_type", specItem.VersionType, pppoeItem.VersionType); err != nil {
				return err
			}
		}
		if specItem.Code != 0 {
			if err := validateUint32ItemField(itemName, "code", specItem.Code, pppoeItem.Code); err != nil {
				return err
			}
		}
		if specItem.SessionId != 0 {
			if err := validateUint32ItemField(itemName, "session_id", specItem.SessionId, pppoeItem.SessionId); err != nil {
				return err
			}
		}
		if specItem.Length != 0 {
			if err := validateUint32ItemField(itemName, "length", specItem.Length, pppoeItem.Length); err != nil {
				return err
			}
		}
	}

	if pppoeItem != nil {
		if pppoeItem.VersionType > 0xFF {
			return fmt.Errorf("invalid version_type (%d), must be equal or lower than 255", pppoeItem.VersionType)
		}
		if pppoeItem.Code > 0xFF {
			return fmt.Errorf("invalid code (%d), must be equal or lower than 255", pppoeItem.Code)
		}
		if pppoeItem.SessionId > 0xFFFF {
			return fmt.Errorf("invalid session_id (%d), must be equal or lower than 65,535", pppoeItem.SessionId)
		}
		if pppoeItem.Length > 0xFFFF {
			return fmt.Errorf("invalid length (%d), must be equal or lower than 65,535", pppoeItem.Length)
		}
	}
	return nil
}

func validateRteFlowItemPppoeProtoId(itemName string, spec, item *any.Any) error {
	specItem := new(flow.RteFlowItemPppoeProtoId)
	if spec != nil {
		if err := ptypes.UnmarshalAny(spec, specItem); err != nil {
			return fmt.Errorf("could not unmarshal spec (%v) %v", spec, err)
		}
	}

	pppoeProtoIdItem := new(flow.RteFlowItemPppoeProtoId)
	if err := ptypes.UnmarshalAny(item, pppoeProtoIdItem); err != nil {
		return fmt.Errorf("could not unmarshal %s: %v", itemName, err)
	}

	if specItem != nil {
		if err := validateUint32ItemField(itemName, "proto_id", specItem.ProtoId, pppoeProtoIdItem.ProtoId); err != nil {
			return err
		}
	}

	if pppoeProtoIdItem != nil {
		if pppoeProtoIdItem.ProtoId > 0xFFFF {
			return fmt.Errorf("invalid proto_id (%d), must be equal or lower than 65,535", pppoeProtoIdItem.ProtoId)
		}
	}

	return nil
}

func validateUint32ItemField(itemName, field string, specValue, fieldValue uint32) error {
	if itemName == "spec" {
		return nil
	}
	if specValue == 0 && fieldValue != 0 {
		return fmt.Errorf("spec.%s must be specified", field)
	}
	if itemName == "last" {
		if fieldValue < specValue {
			return fmt.Errorf("last.%s (%d) must be higher or equal to spec: %s (%d)", field, fieldValue, field, specValue)
		}
	}

	return nil
}
