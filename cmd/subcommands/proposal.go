package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/zzpu/tron-sdk/pkg/address"
	"github.com/zzpu/tron-sdk/pkg/client/transaction"
	"github.com/zzpu/tron-sdk/pkg/common"
	"github.com/zzpu/tron-sdk/pkg/keystore"
	"github.com/zzpu/tron-sdk/pkg/store"
)

var (
	newOnlyProposals = false
)

func proposalSub() []*cobra.Command {
	cmdProposalList := &cobra.Command{
		Use:   "list",
		Short: "List network proposals",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := conn.ProposalsList()
			if err != nil {
				return err
			}

			if noPrettyOutput {
				fmt.Println(list.Proposals)
				return nil
			}

			result := make(map[string]interface{})

			pList := make([]map[string]interface{}, 0)
			for _, proposal := range list.Proposals {
				approvals := make([]string, len(proposal.Approvals))
				for i, a := range proposal.Approvals {
					approvals[i] = address.Address(a).String()
				}
				expired := false
				expiration := time.Unix(proposal.ExpirationTime/1000, 0)
				if expiration.Before(time.Now()) {
					expired = true
					if newOnlyProposals && expired {
						continue
					}
				}

				data := map[string]interface{}{
					"ID":             proposal.ProposalId,
					"Proposer":       address.Address(proposal.ProposerAddress).String(),
					"CreateTime":     time.Unix(proposal.CreateTime/1000, 0),
					"ExpirationTime": expiration,
					"Expired":        expired,
					"Parameters":     proposal.Parameters,
					"Approvals":      approvals,
				}
				pList = append([]map[string]interface{}{data}, pList...)
			}
			result["totalCount"] = len(list.Proposals)
			result["filterCount"] = len(pList)
			result["proposals"] = pList
			asJSON, _ := json.Marshal(result)
			fmt.Println(common.JSONPrettyFormat(string(asJSON)))
			return nil
		},
	}
	cmdProposalList.Flags().BoolVar(&newOnlyProposals, "new", false, "Show only new proposals")

	cmdProposalApprove := &cobra.Command{
		Use:   "approve",
		Short: "Approve network proposal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if signerAddress.String() == "" {
				return fmt.Errorf("no signer specified")
			}

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			confirm, err := strconv.ParseBool(args[1])
			if err != nil {
				return err
			}

			tx, err := conn.ProposalApprove(signerAddress.String(), id, confirm)
			if err != nil {
				return err
			}
			var ctrlr *transaction.Controller
			if useLedgerWallet {
				account := keystore.Account{Address: signerAddress.GetAddress()}
				ctrlr = transaction.NewController(conn, nil, &account, tx.Transaction, opts)
			} else {
				ks, acct, err := store.UnlockedKeystore(signerAddress.String(), passphrase)
				if err != nil {
					return err
				}
				ctrlr = transaction.NewController(conn, ks, acct, tx.Transaction, opts)
			}
			if err = ctrlr.ExecuteTransaction(); err != nil {
				return err
			}

			if noPrettyOutput {
				fmt.Println(tx, ctrlr.Receipt, ctrlr.Result)
				return nil
			}

			result := make(map[string]interface{})
			result["from"] = signerAddress.String()
			result["txID"] = common.BytesToHexString(tx.GetTxid())
			result["blockNumber"] = ctrlr.Receipt.BlockNumber
			result["message"] = string(ctrlr.Result.Message)
			result["receipt"] = map[string]interface{}{
				"fee":      ctrlr.Receipt.Fee,
				"netFee":   ctrlr.Receipt.Receipt.NetFee,
				"netUsage": ctrlr.Receipt.Receipt.NetUsage,
			}

			asJSON, _ := json.Marshal(result)
			fmt.Println(common.JSONPrettyFormat(string(asJSON)))
			return nil
		},
	}

	return []*cobra.Command{cmdProposalList, cmdProposalApprove}
}

func init() {
	cmdProposal := &cobra.Command{
		Use:   "proposal",
		Short: "Network upgrade proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}

	cmdProposal.AddCommand(proposalSub()...)
	RootCmd.AddCommand(cmdProposal)
}
