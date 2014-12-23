package btcassange

import (
	"encoding/hex"
	"fmt"
	"github.com/conformal/btcwire"
	"strings"
	"testing"
)

//coinbase txn f9d077622ef057c50ac6a7d951dc225febab8f5ac5ff71bd2daf4a38ca4a32ed
const sTxn01 string = "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff5303521c0504000175ef07124d696e656420627920425443204775696c642cfabe6d6d251b54903d746842ffa270ec30d85c4c3e02dff64cd319f8c433bbb471deac7401000000000000000800017a833f140000ffffffff0101622995000000001976a91427a1f12771de5cc3b73941664b2537c15316be4388ac00000000"
const txn01_Type int64 = COINBASE
const txn01_TxnHashStr = "f9d077622ef057c50ac6a7d951dc225febab8f5ac5ff71bd2daf4a38ca4a32ed"
const txn01_AddrStr string = "27a1f12771de5cc3b73941664b2537c15316be43"
const txn01_Value int64 = 2502517249

func Test_SplitTx_1(t *testing.T) {
	bTxn01, _ := hex.DecodeString(sTxn01)
	txOutSet, txInSet, err := SplitTx(bTxn01)
	fmt.Println(txOutSet, txInSet)
	if err != nil {
		t.Error(err)
	}
	if len(txOutSet) != 1 {
		t.Error("Tx output count error.")
	}
	if txOutSet[0].Type != COINBASE {
		t.Error("Tx output type error.")
	}
	if hex.EncodeToString(txOutSet[0].Addr) != txn01_AddrStr {
		t.Error("Tx addr error.")
	}
	if txOutSet[0].Value != txn01_Value {
		t.Error("Tx value error.")
	}
	if txOutSet[0].Index != 0 {
		t.Error("Tx index error")
	}
	th, _ := btcwire.NewShaHash(txOutSet[0].TxHash)
	if th.String() != txn01_TxnHashStr {
		t.Error("Tx hash error")
	}
}

//common txn 017aaa8beccb80455bd4da63e7abdf534b286f77ee0a35de1c8faa959b682263
const sTxn02 string = "010000000548f6b1e712a2848014210fc16bd5d183d2930b563d20aa625c9635a625890e01010000006b4830450221009ab57ec53af7900ac1ac980de6302eeb3ac36212e447e05cc09618e785793de402201847f8db2e0793306e28bd27053dec7318876bca039c3e89e5041b66902eed6c012103e05745a1a0dbad51bc852b824b0317600047dcd1dbf6ee98be2d0d212eb246e7ffffffff5e5e6ebd275eaeb2235714586e6d4cfb903b995441e29dbd65ec5b2bed9f0b17010000006b48304502210081e1868f79bf843251cbe1a49bf688167f36d1c83e1243cdeab6bcbf20a6faea022036dfbace72801aff1c74037f6b1e1bfc98412ddb4ec8c0902743dbab158d7aa80121026a4b7fdb8f812f53c703cee752012e4ac198cb062a1c27d83341dab1b1bb93f2ffffffff49579735b2df610e37632a7a58a55762f8ba91305fe7f1862333c7c889773825010000006b483045022100f5f1f19d66149582a432d12ad01db0f49195b34a6db0a54afe526f53cd9ed98b022079b2ade77d4e1f2fa0555bb34597a3d81bd8adeaec2de83bbec17d1a79153c320121022f9d52dcb1520f777a3a97728cf5793d25a62e3353152243ff74f1765881b386ffffffff712aeb6360309deca980437d810bf00fae1bdc15e68f57da0398a9a817d079fa010000006a473044022014d0a4a564e915a26f30ec9feb4e748c146e9c273797e186c3d6fd5d65838b9c02202e60418b25867ba464a9dcb490648ad200eef59922678c2ffa5683ac6248eb520121028f75eededcb533f4ecb47221f322521766a88fcdd7c5cad760eb679477e1505cffffffff70e2fb5ee861814a822c7fb661e5a6020e62212836a194cd0472c9755e3eb55a010000006a4730440220734a8ac21d0336b9909c3d50bc1af95a6be2aec5e09ede64609529c567492e4e02205cec9f3a5656989010bbfd8d6d09a38383d5e2948370ab18f82025d6ed2cd78c01210267003c7fd21dfc4cacb8541a9f205f2b96d367d0e4c6be3d0897e07d2f38e1c6ffffffff02f064b900000000001976a914bd5086ab2eda5aee18102c7997e047af06dbc35288ac0049d971020000001976a914f0f711f6a225d2e1c85d13116791bee53a9a1b0b88ac00000000"
const txn02InCnt int = 5

var txn02Out [2]map[string]interface{} = [2]map[string]interface{}{
	{"addr": "BD5086AB2EDA5AEE18102C7997E047AF06DBC352", "type": COMMONOUT, "value": 12150000, "index": 0, "hash": "017aaa8beccb80455bd4da63e7abdf534b286f77ee0a35de1c8faa959b682263"},
	{"addr": "F0F711F6A225D2E1C85D13116791BEE53A9A1B0B", "type": COMMONOUT, "value": 10500000000, "index": 1, "hash": "017aaa8beccb80455bd4da63e7abdf534b286f77ee0a35de1c8faa959b682263"},
}
var txn02In [5]map[string]interface{} = [5]map[string]interface{}{
	{"hash": "017aaa8beccb80455bd4da63e7abdf534b286f77ee0a35de1c8faa959b682263", "prehash": "010e8925a635965c62aa203d560b93d283d1d56bc10f21148084a212e7b1f648", "preindex": 1},
	{"hash": "017aaa8beccb80455bd4da63e7abdf534b286f77ee0a35de1c8faa959b682263", "prehash": "170b9fed2b5bec65bd9de24154993b90fb4c6d6e58145723b2ae5e27bd6e5e5e", "preindex": 1},
	{"hash": "017aaa8beccb80455bd4da63e7abdf534b286f77ee0a35de1c8faa959b682263", "prehash": "25387789c8c7332386f1e75f3091baf86257a5587a2a63370e61dfb235975749", "preindex": 1},
	{"hash": "017aaa8beccb80455bd4da63e7abdf534b286f77ee0a35de1c8faa959b682263", "prehash": "fa79d017a8a99803da578fe615dc1bae0ff00b817d4380a9ec9d306063eb2a71", "preindex": 1},
	{"hash": "017aaa8beccb80455bd4da63e7abdf534b286f77ee0a35de1c8faa959b682263", "prehash": "5ab53e5e75c97204cd94a1362821620e02a6e561b67f2c824a8161e85efbe270", "preindex": 1},
}

func TestSplitTx_2(t *testing.T) {
	bTxn02, _ := hex.DecodeString(sTxn02)
	txOutSet, txInSet, err := SplitTx(bTxn02)
	if err != nil {
		t.Error(err)
	}
	if len(txOutSet) != len(txn02Out) {
		t.Error("Tx output count error.")
	}
	for i := range txOutSet {
		if hex.EncodeToString(txOutSet[i].Addr) != strings.ToLower(txn02Out[i]["addr"].(string)) {
			t.Errorf("Tx output %d address error.", i)
		}
		if txOutSet[i].Type != int64(txn02Out[i]["type"].(int64)) {
			t.Errorf("Tx output %d type error.", i)
		}
		if txOutSet[i].Value != int64(txn02Out[i]["value"].(int)) {
			t.Errorf("Tx output %d value error.", i)
		}
		if txOutSet[i].Index != int64(txn02Out[i]["index"].(int)) {
			t.Errorf("Tx output %d index error.", i)
		}
		th, _ := btcwire.NewShaHash(txOutSet[i].TxHash)
		if th.String() != txn02Out[i]["hash"] {
			t.Error("Tx output %d hash error.", i)
		}

	}
	if len(txInSet) != len(txn02In) {
		t.Error("Tx input count error.")
	}
	for i := range txInSet {
		th, _ := btcwire.NewShaHash(txInSet[i].TxHash)
		if th.String() != txn02In[i]["hash"] {
			t.Error("Tx input %d hash error.", i)
		}
		pth, _ := btcwire.NewShaHash(txInSet[i].PreTxHash)
		if pth.String() != txn02In[i]["prehash"] {
			t.Error("Tx input %d prehash error.", i)
		}
		if txInSet[i].PreOutIndex != int64(txn02In[i]["preindex"].(int)) {
			t.Error("Tx input %d preindex error.", i)
		}
	}

}
