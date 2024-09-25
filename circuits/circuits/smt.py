from poseidon_py.poseidon_hash import (
    poseidon_hash,
    poseidon_hash_single,
    poseidon_hash_many as poseidon
)


# SMTHash1 and SMTHash2
def smt_hash1(key, value):
    print(poseidon([1, key, value]))
    return poseidon([1, key, value])

def smt_hash2(left, right):
    print(poseidon([left, right]))
    return poseidon([left, right])

# SMTLevIns
def smt_lev_ins(siblings, enabled):
    n_levels = len(siblings)
    lev_ins = [0] * n_levels
    prev = enabled
    for i in range(n_levels):
        lev_ins[i] = prev * (1 - (siblings[i] == 0))
        prev = lev_ins[i]
    return lev_ins

# SMTProcessorSM
def smt_processor_sm(prev_state, is0, xor, fnc, lev_ins):
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
def smt_processor_level(state, sibling, old1leaf, new1leaf, newlrbit, old_child, new_child):
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
def smt_processor(old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc):
    n_levels = len(siblings)
    enabled = fnc[0] + fnc[1] - fnc[0] * fnc[1]

    hash1_old = smt_hash1(old_key, old_value)
    hash1_new = smt_hash1(new_key, new_value)

    n2b_old = bin(old_key)[2:].zfill(n_levels)
    n2b_new = bin(new_key)[2:].zfill(n_levels)

    lev_ins = smt_lev_ins(siblings, enabled)

    xors = [int(a) ^ int(b) for a, b in zip(n2b_old, n2b_new)]

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
        sm[i] = smt_processor_sm(prev_state, is_old0, xors[i], fnc, lev_ins[i])

    levels = [{}] * n_levels
    for i in range(n_levels - 1, -1, -1):
        old_child = 0 if i == n_levels - 1 else levels[i+1]['oldRoot']
        new_child = 0 if i == n_levels - 1 else levels[i+1]['newRoot']
        levels[i] = smt_processor_level(sm[i], siblings[i], hash1_old, hash1_new, int(n2b_new[i]), old_child, new_child)

    top_switcher_sel = fnc[0] * fnc[1]
    top_switcher_l = levels[0]['oldRoot']
    top_switcher_r = levels[0]['newRoot']

    new_root = enabled * (top_switcher_r - old_root) + old_root

    are_keys_equal = old_key == new_key
    keys_ok = (1 - fnc[0]) * fnc[1] * (1 - are_keys_equal)

    assert keys_ok == 0, "Keys do not match for update operation"

    return new_root

# Example usage
old_root = 0  # Initial root hash
siblings = [0, 0]  # Example with 2 levels
old_key = 1
old_value = 0
is_old0 = True
new_key = 1
new_value = 10
fnc = [1, 0]  # Insert

new_root = smt_processor(old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc)
print(f"New root after insertion: {new_root}")

# Update
old_root = new_root
old_value = 10
new_value = 20
fnc = [0, 1]  # Update

new_root = smt_processor(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
print(f"New root after update: {new_root}")

# Delete
old_root = new_root
new_value = 0
fnc = [1, 1]  # Delete

new_root = smt_processor(old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
print(f"New root after deletion: {new_root}")