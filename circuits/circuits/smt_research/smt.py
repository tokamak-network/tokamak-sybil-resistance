from py_ecc.fields import bn128_FQ as FQ
from py_ecc import bn128
from poseidon import Poseidon

class FR(FQ):
    field_modulus = bn128.curve_order

###################################################
#Utility functions
###################################################
def Num2Bits_strict(n, bits):
    return [int(b) for b in bin(int(n))[2:].zfill(bits)]

def XOR(a, b):
    return FR(a + b - 2*a*b)

def IsEqual(a, b):
    return FR(1) if a == b else FR(0)

def AND(a, b):
    return FR(a * b)

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
    return sel * R + (FR(1) - sel) * L

def ForceEqualIfEnabled(enabled, in1, in2):
    return enabled * (in1 - in2)

###################################################
#SMT Circuits: https://github.com/iden3/circomlib/tree/master/circuits/smt
###################################################

# SMTHash1 and SMTHash2
def SMTHash1(key, value):
    return Poseidon([1, key, value])

def SMTHash2(L, R):
    return Poseidon([L, R])

# SMTLevIns
def SMTLevIns(siblings, enabled):
    n_levels = len(siblings)
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
def SMTProcessorSM(prev_state, is0, xor, fnc, lev_ins):
    # Input state
    prev_top = prev_state['prev_top']
    prev_old0 = prev_state['prev_old0']
    prev_bot = prev_state['prev_bot']
    prev_new1 = prev_state['prev_new1']
    prev_na = prev_state['prev_na']
    prev_upd = prev_state['prev_upd']

    # Auxiliary calculation
    aux1 = prev_top * lev_ins
    aux2 = aux1 * fnc[0]

    # State calculation
    st_top = prev_top - aux1
    st_old0 = aux2 * is0
    st_new1 = (aux2 - st_old0 + prev_bot) * xor
    st_bot = (FR(1) - xor) * (aux2 - st_old0 + prev_bot)
    st_upd = aux1 - aux2
    st_na = prev_new1 + prev_old0 + prev_na + prev_upd

    # Return result
    return {
        'st_top': st_top,
        'st_old0': st_old0,
        'st_bot': st_bot,
        'st_new1': st_new1,
        'st_na': st_na,
        'st_upd': st_upd
    }

# SMTProcessorLevel
def SMTProcessorLevel(state, sibling, old1leaf, new1leaf, newlrbit, old_child, new_child):
    old_root = (state['st_top'] * sibling + 
                state['st_old0'] * old_child + 
                state['st_bot'] * old1leaf + 
                state['st_na'] * old_child)

    new_root = (state['st_top'] * sibling + 
                state['st_old0'] * new_child + 
                state['st_bot'] * ((FR(1) - newlrbit) * old1leaf + newlrbit * new1leaf) + 
                state['st_new1'] * ((FR(1) - newlrbit) * new1leaf + newlrbit * old1leaf) + 
                state['st_na'] * new_child + 
                state['st_upd'] * new1leaf)

    return {
        'oldRoot': old_root,
        'newRoot': new_root
    }

# SMTProcessor
def SMTProcessor(old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc):
    n_levels = len(siblings)
    enabled = fnc[0] + fnc[1] - fnc[0] * fnc[1]

    hash1_old = SMTHash1(old_key, old_value)
    hash1_new = SMTHash1(new_key, new_value)

    n2b_old = Num2Bits_strict(old_key, n_levels)
    n2b_new = Num2Bits_strict(new_key, n_levels)
    print("old_key", old_key)
    print("new_key", new_key)
    print("n2b_old", n2b_old)
    print("n2b_new", n2b_new)

    lev_ins = SMTLevIns(siblings, enabled)

    xors = [XOR(FR(a), FR(b)) for a, b in zip(n2b_old, n2b_new)]
    print("xors", xors)

    sm = [{}] * n_levels
    for i in range(n_levels):
        prev_state = {
            'prev_top': enabled if i == 0 else sm[i-1]['st_top'],
            'prev_old0': FR(0) if i == 0 else sm[i-1]['st_old0'],
            'prev_bot': FR(0) if i == 0 else sm[i-1]['st_bot'],
            'prev_new1': FR(0) if i == 0 else sm[i-1]['st_new1'],
            'prev_na': FR(1) - enabled if i == 0 else sm[i-1]['st_na'],
            'prev_upd': FR(0) if i == 0 else sm[i-1]['st_upd']
        }
        sm[i] = SMTProcessorSM(prev_state, FR(is_old0), xors[i], fnc, lev_ins[i])

    levels = [{}] * n_levels
    for i in range(n_levels - 1, -1, -1):
        old_child = FR(0) if i == n_levels - 1 else levels[i+1]['oldRoot']
        new_child = FR(0) if i == n_levels - 1 else levels[i+1]['newRoot']
        levels[i] = SMTProcessorLevel(sm[i], siblings[i], hash1_old, hash1_new, FR(n2b_new[i]), old_child, new_child)

    top_switcher_sel = fnc[0] * fnc[1]
    top_switcher_l = levels[0]['oldRoot']
    top_switcher_r = levels[0]['newRoot']

    new_root = Switcher(enabled, old_root, Switcher(top_switcher_sel, top_switcher_l, top_switcher_r))

    are_keys_equal = IsEqual(old_key, new_key)
    keys_ok = MultiAND([FR(1) - fnc[0], fnc[1], FR(1) - are_keys_equal])

    assert keys_ok == FR(0), "Keys do not match for update operation"

    print("sm", sm)
    total_state = sm[n_levels-1]['st_na'] + sm[n_levels-1]['st_new1'] + sm[n_levels-1]['st_old0'] + sm[n_levels-1]['st_upd']

    assert total_state == FR(1), f"Invalid state at the last level: {total_state}"

    ForceEqualIfEnabled(enabled, old_root, top_switcher_l)

    return new_root

###################################################
# Example usage
###################################################
old_root = 0  # Initial root hash
siblings = [0, 0]  # Example with 2 levels
old_key = 1
old_value = 0
is_old0 = True
new_key = 1
new_value = 10
fnc = [1, 0]  # Insert

new_root = SMTProcessor(old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc)
print(f"New root after insertion: {new_root}")

# Update
old_root = new_root
old_value = 10
new_value = 20
fnc = [0, 1]  # Update

new_root = SMTProcessor(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
print(f"New root after update: {new_root}")

# Delete
old_root = new_root
new_value = 0
fnc = [1, 1]  # Delete

new_root = SMTProcessor(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
print(f"New root after deletion: {new_root}")

z = SMTHash2(1, 2)
print(z)