from py_ecc.fields import bn128_FQ as FQ
from py_ecc import bn128
from poseidon import Poseidon

class FR(FQ):
    field_modulus = bn128.curve_order

###################################################
#Utility functions
###################################################
def Num2Bits_strict(in_value):
    n = int(in_value)
    out = [0] * 254 #Num2Bits_strict output is 254bit array
    if n >= (1 << 254):
        raise ValueError("Input is too large for 254 bits")
    for i in range(254):
        out[i] = (n >> i) & 1
    return out

def XOR(a, b):
    return a + b - 2*a*b

def IsEqual(a, b):
    return 1 if a == b else 0

def AND(a, b):
    return a * b

def MultiAND(inputs):
    n = len(inputs)
    
    if n == 1:
        return inputs[0]
    elif n == 2:
        return AND(inputs[0], inputs[1])
    else:
        n1 = n // 2
        n2 = n - n1
        
        left_result = MultiAND(inputs[:n1])
        right_result = MultiAND(inputs[n1:])
        
        return AND(left_result, right_result)

def Switcher(sel, L, R):
    aux = (R - L) * sel
    outL = aux + L
    outR = -aux + R
    return outL, outR

def ForceEqualIfEnabled(enabled, in1, in2):
    return enabled * (in1 - in2)

###################################################
#SMT Circuits: https://github.com/iden3/circomlib/tree/master/circuits/smt
###################################################

# SMTHash1 and SMTHash2
def SMTHash1(key, value):
    return Poseidon([int(key), int(value), 1])

def SMTHash2(L, R):
    return Poseidon([int(L), int(R)])

# SMTLevIns
def SMTLevIns(n_levels, siblings, enabled):
    lev_ins = [FR(0)] * n_levels
    done = [FR(0)] * (n_levels - 1)

    # Check last level
    if enabled:
        assert IsEqual(siblings[-1], FR(0)), "Last sibling must be zero if enabled"

    # Calculate from highest to lowest level
    lev_ins[-1] = FR(1) - IsEqual(siblings[-2], FR(0))
    done[-1] = lev_ins[-1]

    for i in range(n_levels - 2, 0, -1):
        lev_ins[i] = (FR(1) - done[i]) * (FR(1) - IsEqual(siblings[i-1], FR(0)))
        done[i-1] = lev_ins[i] + done[i]

    lev_ins[0] = FR(1) - done[0]

    return lev_ins

# SMTProcessorSM
def SMTProcessorSM(xor, is0, levIns, fnc, prev_top, prev_old0, prev_bot, prev_new1, prev_na, prev_upd):
    aux1 = prev_top * levIns
    aux2 = aux1 * fnc[0]

    st_top = prev_top - aux1

    st_old0 = aux2 * is0

    st_new1 = (aux2 - st_old0 + prev_bot) * xor

    st_bot = (FR(1) - xor) * (aux2 - st_old0 + prev_bot)

    st_upd = aux1 - aux2

    st_na = prev_new1 + prev_old0 + prev_na + prev_upd

    return {
        'st_top': st_top,
        'st_old0': st_old0,
        'st_bot': st_bot,
        'st_new1': st_new1,
        'st_na': st_na,
        'st_upd': st_upd
    }

# SMTProcessorLevel
def SMTProcessorLevel(st_top, st_old0, st_bot, st_new1, st_na, st_upd,
                      sibling, old1leaf, new1leaf, newlrbit, oldChild, newChild):
    aux = [0] * 4

    # Old side
    oldSwitcher_L, oldSwitcher_R = Switcher(newlrbit, oldChild, sibling)
    #print(f"oldSwitcher_L: {oldSwitcher_L}, oldSwitcher_R: {oldSwitcher_R}")
    oldProofHash = SMTHash2(oldSwitcher_L, oldSwitcher_R)

    #print(f"oldProofHash: {oldProofHash}")

    aux[0] = old1leaf * (st_bot + st_new1 + st_upd)
    oldRoot = aux[0] + oldProofHash * st_top

    # New side
    aux[1] = newChild * (st_top + st_bot)
    newSwitcher_L = aux[1] + new1leaf * st_new1

    aux[2] = sibling * st_top
    newSwitcher_R = aux[2] + old1leaf * st_new1

    newSwitcher_outL, newSwitcher_outR = Switcher(newlrbit, newSwitcher_L, newSwitcher_R)
    newProofHash = SMTHash2(newSwitcher_outL, newSwitcher_outR)
    #print(f"newProofHash: {newProofHash}")
    aux[3] = newProofHash * (st_top + st_bot + st_new1)
    newRoot = aux[3] + new1leaf * (st_old0 + st_upd)
    #print(f"newRoot: {newRoot}")
    return {'oldRoot': oldRoot, 'newRoot': newRoot}


# SMTProcessor
def SMTProcessor(nLevels, oldRoot, siblings, oldKey, oldValue, isOld0, newKey, newValue, fnc):
    enabled = fnc[0] + fnc[1] - fnc[0] * fnc[1]

    hash1Old = SMTHash1(oldKey, oldValue)
    hash1New = SMTHash1(newKey, newValue)

    n2bOld = Num2Bits_strict(oldKey)
    n2bNew = Num2Bits_strict(newKey)

    smtLevIns = SMTLevIns(nLevels, siblings, enabled)

    xors = [XOR(n2bOld[i], n2bNew[i]) for i in range(nLevels)]

    sm = []
    for i in range(nLevels):
        if i == 0:
            prev_top = enabled
            prev_old0 = FR(0)
            prev_bot = FR(0)
            prev_new1 = FR(0)
            prev_na = FR(1) - enabled
            prev_upd = FR(0)
        else:
            prev_top = sm[i-1]['st_top']
            prev_old0 = sm[i-1]['st_old0']
            prev_bot = sm[i-1]['st_bot']
            prev_new1 = sm[i-1]['st_new1']
            prev_na = sm[i-1]['st_na']
            prev_upd = sm[i-1]['st_upd']

        sm.append(SMTProcessorSM(
            xors[i], isOld0, smtLevIns[i], fnc,
            prev_top, prev_old0, prev_bot, prev_new1, prev_na, prev_upd
        ))

    assert sm[nLevels-1]['st_na'] + sm[nLevels-1]['st_new1'] + sm[nLevels-1]['st_old0'] + sm[nLevels-1]['st_upd'] == FR(1)

    levels = [None] * nLevels
    for i in range(nLevels - 1, -1, -1):
        level = SMTProcessorLevel(
            sm[i]['st_top'], sm[i]['st_old0'], sm[i]['st_bot'], sm[i]['st_new1'], sm[i]['st_na'], sm[i]['st_upd'],
            siblings[i], hash1Old, hash1New, n2bNew[i],
            FR(0) if i == nLevels - 1 else levels[i+1]['oldRoot'],
            FR(0) if i == nLevels - 1 else levels[i+1]['newRoot']
        )
        levels[i] = level

    #print(f"levels: {levels}")

    topSwitcher_L, topSwitcher_R = Switcher(fnc[0] * fnc[1],levels[0]['oldRoot'], levels[0]['newRoot'])

    #print(f"topSwitcher_L: {topSwitcher_L}, topSwitcher_R: {topSwitcher_R}")


    # print(f"oldRoot: {oldRoot}, type: {type(oldRoot)}, class: {oldRoot.__class__.__name__}")
    # print(f"topSwitcher['outL']: {topSwitcher['outL']}, type: {type(topSwitcher['outL'])}, class: {topSwitcher['outL'].__class__.__name__}")

    checkOldInput = ForceEqualIfEnabled(enabled, oldRoot, topSwitcher_L)

    #print("oldRoot", oldRoot)

    newRoot = enabled * (topSwitcher_R - oldRoot) + oldRoot
    #print(f"newRoot calculation: enabled={enabled}, topSwitcher_R={topSwitcher_R}, oldRoot={oldRoot}")
    #print(f"newRoot: {newRoot}")

    # Check keys are equal if updating
    areKeyEquals = IsEqual(oldKey, newKey)
    keysOk = MultiAND([FR(1) - fnc[0], fnc[1], FR(1) - areKeyEquals])

    assert keysOk == FR(0)

    return newRoot

###################################################
# Example usage
###################################################
# fnc = [1, 0]  # Insert
# old_root = 0  # Initial root hash
# siblings = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0]  # Example with 2 levels
# nlevels = len(siblings)
# old_key = 0
# old_value = 0
# is_old0 = 1
# new_key = 111
# new_value = 222

# #new_root should be 9308772482099879945566979599408036177864352098141198065063141880905857869998
# new_root = SMTProcessor(nlevels, old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc)
# print(f"New root after insertion: {new_root}")

# # Update
# old_root = new_root
# old_value = 10
# new_value = 20
# fnc = [0, 1]  # Update

# new_root = SMTProcessor(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
# print(f"New root after update: {new_root}")

# # Delete
# old_root = new_root
# new_value = 0
# fnc = [1, 1]  # Delete

# new_root = SMTProcessor(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
# print(f"New root after deletion: {new_root}")

# z = SMTHash2(0, 0)
# print(z)