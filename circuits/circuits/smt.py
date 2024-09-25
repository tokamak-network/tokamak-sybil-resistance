from poseidon_py.poseidon_hash import poseidon_hash_many as poseidon
#Todo: use bn254 field


# Utility functions
def Num2Bits_strict(n, bits):
    return [int(b) for b in bin(n)[2:].zfill(bits)]

def XOR(a, b):
    return a + b - 2*a*b

def IsEqual(a, b):
    return int(a == b)

def AND(a, b):
    return a * b

def MultiAND(*args):
    n = len(args)
    if n == 1:
        return args[0]
    elif n == 2:
        return AND(args[0], args[1])
    else:
        n1 = n // 2
        n2 = n - n1
        return AND(MultiAND(*args[:n1]), MultiAND(*args[n1:]))

def Switcher(sel, L, R):
    return sel * R + (1 - sel) * L

def ForceEqualIfEnabled(enabled, in1, in2):
    return enabled * (in1 - in2)

# SMTHash1 and SMTHash2
def SMTHash1(key, value):
    return poseidon([1, key, value])

def SMTHash2(L, R):
    return poseidon([L, R])

# SMTLevIns
def SMTLevIns(siblings, enabled):
    n_levels = len(siblings)
    lev_ins = [0] * n_levels
    prev = enabled
    for i in range(n_levels):
        lev_ins[i] = prev * (1 - IsEqual(siblings[i], 0))
        prev = lev_ins[i]
    return lev_ins

# SMTProcessorSM
def SMTProcessorSM(prev_state, is0, xor, fnc, lev_ins):
    st_top = prev_state['prev_top'] * (1 - xor)
    st_old0 = prev_state['prev_top'] * xor * is0 + prev_state['prev_old0'] * (1 - xor)
    st_bot = (prev_state['prev_top'] * xor * (1 - is0) + prev_state['prev_old0'] * xor + 
              prev_state['prev_bot'] * (1 - xor))
    st_new1 = prev_state['prev_new1'] + lev_ins * (1 - fnc[1])
    st_na = prev_state['prev_na']
    st_upd = prev_state['prev_upd'] + lev_ins * fnc[1] * (1 - fnc[0])

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
                state['st_bot'] * ((1 - newlrbit) * old1leaf + newlrbit * new1leaf) + 
                state['st_new1'] * ((1 - newlrbit) * new1leaf + newlrbit * old1leaf) + 
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

    lev_ins = SMTLevIns(siblings, enabled)

    xors = [XOR(a, b) for a, b in zip(n2b_old, n2b_new)]

    sm = [{}] * n_levels
    for i in range(n_levels):
        prev_state = {
            'prev_top': enabled if i == 0 else sm[i-1]['st_top'],
            'prev_old0': 0 if i == 0 else sm[i-1]['st_old0'],
            'prev_bot': 0 if i == 0 else sm[i-1]['st_bot'],
            'prev_new1': 0 if i == 0 else sm[i-1]['st_new1'],
            'prev_na': 1 - enabled if i == 0 else sm[i-1]['st_na'],
            'prev_upd': 0 if i == 0 else sm[i-1]['st_upd']
        }
        sm[i] = SMTProcessorSM(prev_state, is_old0, xors[i], fnc, lev_ins[i])

    levels = [{}] * n_levels
    for i in range(n_levels - 1, -1, -1):
        old_child = 0 if i == n_levels - 1 else levels[i+1]['oldRoot']
        new_child = 0 if i == n_levels - 1 else levels[i+1]['newRoot']
        levels[i] = SMTProcessorLevel(sm[i], siblings[i], hash1_old, hash1_new, n2b_new[i], old_child, new_child)

    top_switcher_sel = fnc[0] * fnc[1]
    top_switcher_l = levels[0]['oldRoot']
    top_switcher_r = levels[0]['newRoot']

    new_root = Switcher(enabled, old_root, Switcher(top_switcher_sel, top_switcher_l, top_switcher_r))

    are_keys_equal = IsEqual(old_key, new_key)
    keys_ok = MultiAND(1 - fnc[0], fnc[1], 1 - are_keys_equal)

    assert keys_ok == 0, "Keys do not match for update operation"
    assert sm[n_levels-1]['st_na'] + sm[n_levels-1]['st_new1'] + sm[n_levels-1]['st_old0'] + sm[n_levels-1]['st_upd'] == 1, "Invalid state at the last level"

    ForceEqualIfEnabled(enabled, old_root, top_switcher_l)

    return new_root

# Input validation
def validate_inputs(old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc):
    assert 0 <= old_root < 2**256, "Invalid old_root"
    assert all(0 <= s < 2**256 for s in siblings), "Invalid siblings"
    assert 0 <= old_key < 2**256, "Invalid old_key"
    assert 0 <= old_value < 2**256, "Invalid old_value"
    assert isinstance(is_old0, bool), "is_old0 must be boolean"
    assert 0 <= new_key < 2**256, "Invalid new_key"
    assert 0 <= new_value < 2**256, "Invalid new_value"
    assert len(fnc) == 2 and all(f in [0, 1] for f in fnc), "Invalid fnc"

# Example usage
old_root = 0  # Initial root hash
siblings = [0, 0]  # Example with 2 levels
old_key = 1
old_value = 0
is_old0 = True
new_key = 1
new_value = 10
fnc = [1, 0]  # Insert

validate_inputs(old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc)
new_root = SMTProcessor(old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc)
print(f"New root after insertion: {new_root}")

# Update
old_root = new_root
old_value = 10
new_value = 20
fnc = [0, 1]  # Update

validate_inputs(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
new_root = SMTProcessor(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
print(f"New root after update: {new_root}")

# Delete
old_root = new_root
new_value = 0
fnc = [1, 1]  # Delete

validate_inputs(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
new_root = SMTProcessor(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
print(f"New root after deletion: {new_root}")