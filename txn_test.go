package Assange

import (
	"encoding/hex"
	"fmt"
	"github.com/conformal/btcwire"
	"strings"
	"testing"
)

//coinbase txn f9d077622ef057c50ac6a7d951dc225febab8f5ac5ff71bd2daf4a38ca4a32ed
const sTxn01 string = "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff5303521c0504000175ef07124d696e656420627920425443204775696c642cfabe6d6d251b54903d746842ffa270ec30d85c4c3e02dff64cd319f8c433bbb471deac7401000000000000000800017a833f140000ffffffff0101622995000000001976a91427a1f12771de5cc3b73941664b2537c15316be4388ac00000000"
const txn01_Type int8 = 3
const txn01_TxnHashStr = "f9d077622ef057c50ac6a7d951dc225febab8f5ac5ff71bd2daf4a38ca4a32ed"
const txn01_AddrStr string = "27a1f12771de5cc3b73941664b2537c15316be43"
const txn01_Value int64 = 2502517249

//common txn 017aaa8beccb80455bd4da63e7abdf534b286f77ee0a35de1c8faa959b682263
const sTxn02 string = "010000000548f6b1e712a2848014210fc16bd5d183d2930b563d20aa625c9635a625890e01010000006b4830450221009ab57ec53af7900ac1ac980de6302eeb3ac36212e447e05cc09618e785793de402201847f8db2e0793306e28bd27053dec7318876bca039c3e89e5041b66902eed6c012103e05745a1a0dbad51bc852b824b0317600047dcd1dbf6ee98be2d0d212eb246e7ffffffff5e5e6ebd275eaeb2235714586e6d4cfb903b995441e29dbd65ec5b2bed9f0b17010000006b48304502210081e1868f79bf843251cbe1a49bf688167f36d1c83e1243cdeab6bcbf20a6faea022036dfbace72801aff1c74037f6b1e1bfc98412ddb4ec8c0902743dbab158d7aa80121026a4b7fdb8f812f53c703cee752012e4ac198cb062a1c27d83341dab1b1bb93f2ffffffff49579735b2df610e37632a7a58a55762f8ba91305fe7f1862333c7c889773825010000006b483045022100f5f1f19d66149582a432d12ad01db0f49195b34a6db0a54afe526f53cd9ed98b022079b2ade77d4e1f2fa0555bb34597a3d81bd8adeaec2de83bbec17d1a79153c320121022f9d52dcb1520f777a3a97728cf5793d25a62e3353152243ff74f1765881b386ffffffff712aeb6360309deca980437d810bf00fae1bdc15e68f57da0398a9a817d079fa010000006a473044022014d0a4a564e915a26f30ec9feb4e748c146e9c273797e186c3d6fd5d65838b9c02202e60418b25867ba464a9dcb490648ad200eef59922678c2ffa5683ac6248eb520121028f75eededcb533f4ecb47221f322521766a88fcdd7c5cad760eb679477e1505cffffffff70e2fb5ee861814a822c7fb661e5a6020e62212836a194cd0472c9755e3eb55a010000006a4730440220734a8ac21d0336b9909c3d50bc1af95a6be2aec5e09ede64609529c567492e4e02205cec9f3a5656989010bbfd8d6d09a38383d5e2948370ab18f82025d6ed2cd78c01210267003c7fd21dfc4cacb8541a9f205f2b96d367d0e4c6be3d0897e07d2f38e1c6ffffffff02f064b900000000001976a914bd5086ab2eda5aee18102c7997e047af06dbc35288ac0049d971020000001976a914f0f711f6a225d2e1c85d13116791bee53a9a1b0b88ac00000000"
const txn02_itemCnt int = 7

var txn02_addresses [7]string = [7]string{"e1098b09126b747283a46e2e068bcb819f0922e3",
	"E158CEB7B97EA0C9F33A716D52BF0D7A1D1A343B",
	"BC2E5E6AB6B8DDA0D2AF94CC477F0D7047D55425",
	"8568D32C017E3D6C19AF895CDA6BF2F529ABB797",
	"1796DD10EF5B96CE6BD60F98BFAAF9937D979F86",
	"BD5086AB2EDA5AEE18102C7997E047AF06DBC352",
	"F0F711F6A225D2E1C85D13116791BEE53A9A1B0B",
}

func Test_SplitTxnFromString_1(t *testing.T) {
	bTxn01, _ := hex.DecodeString(sTxn01)
	txn_set := SplitTxnFromString(bTxn01)
	println(len(txn_set))
	if len(txn_set) != 1 {
		t.Error("TxItem count Error")
	}

	if txn_set[0].Type != txn01_Type {
		t.Error("Type Error")
	}

	th, _ := btcwire.NewShaHash(txn_set[0].TxnHash)
	if th.String() != txn01_TxnHashStr {
		t.Error("TxnHash Error")
	}

	addr := hex.EncodeToString(txn_set[0].Addr)
	if addr != txn01_AddrStr {
		fmt.Println(addr, txn01_AddrStr)
		t.Error("Addr Error")
	}
}

func Test_SplitTxnFromString_2(t *testing.T) {
	bTxn02, _ := hex.DecodeString(sTxn02)
	txn_set := SplitTxnFromString(bTxn02)
	if len(txn_set) != txn02_itemCnt {
		fmt.Printf("Actual count:%d Expected count:%d\n", len(txn_set), txn02_itemCnt)
		t.Error("TxItem count Error")
	}
	for i := range txn_set {
		if hex.EncodeToString(txn_set[i].Addr) != strings.ToLower(txn02_addresses[i]) {
			t.Errorf("TxItem Address%d Error", i)
		}
	}
}

var bSigScript01 [3]string = [3]string{"4830450221009ab57ec53af7900ac1ac980de6302eeb3ac36212e447e05cc09618e785793de402201847f8db2e0793306e28bd27053dec7318876bca039c3e89e5041b66902eed6c012103e05745a1a0dbad51bc852b824b0317600047dcd1dbf6ee98be2d0d212eb246e7", "30450221009ab57ec53af7900ac1ac980de6302eeb3ac36212e447e05cc09618e785793de402201847f8db2e0793306e28bd27053dec7318876bca039c3e89e5041b66902eed6c01", "03e05745a1a0dbad51bc852b824b0317600047dcd1dbf6ee98be2d0d212eb246e7"}

func Test_DecodeSigScript_1(t *testing.T) {
	sSigScript, _ := hex.DecodeString(bSigScript01[0])
	sig, pk, err := DecodeSigScript(sSigScript)
	if err != nil {
		t.Error(err)
	}
	if hex.EncodeToString(sig) != bSigScript01[1] {
		fmt.Printf("Actual Sig:%s\n", hex.EncodeToString(sig))
		fmt.Printf("Expected Sig:%s\n", bSigScript01[1])
		t.Error("Signature Error")
	}
	if hex.EncodeToString(pk) != bSigScript01[2] {
		fmt.Printf("Actual Pubkey:%s\n", hex.EncodeToString(pk))
		fmt.Printf("Expected Pubkey:%s\n", bSigScript01[2])
		t.Error("Pubkey Error")
	}
}
