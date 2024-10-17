from py_ecc.fields import bn128_FQ as FQ
from py_ecc import bn128
from poseidon import Poseidon

class FR(FQ):
    field_modulus = bn128.curve_order

###################################################
#Utility functions
###################################################

#Number to bits
def Num2Bits_strict(in_value):
    n = int(in_value)
    out = [0] * 254 #Num2Bits_strict output is 254bit array
    if n >= (1 << 254):
        raise ValueError("Input is too large for 254 bits")
    for i in range(254):
        out[i] = (n >> i) & 1
    return out

#XOR for bitwise
def XOR(a, b):
    return a + b - 2*a*b

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

#if sel == 0 then outL = L and outR=R
#if sel == 1 then outL = R and outR=L
def Switcher(sel, L, R):
    aux = (R - L) * sel
    outL = aux + L
    outR = -aux + R
    return outL, outR

def ForceEqualIfEnabled(enabled, in1, in2):
    assert enabled * (in1 - in2) == 0, "ForceEqualIfEnabled failed"

###################################################
#SMT Circuits: https://github.com/iden3/circomlib/tree/master/circuits/smt
###################################################

# SMTHash1 and SMTHash2
def SMTHash1(key, value):
    """
    Computes the hash of a leaf node in the Sparse Merkle Tree.
    Includes a constant '1' for domain separation to distinguish
    leaf node hashes from internal node hashes, preventing collisions.
    """
    return Poseidon([int(key), int(value), 1])

def SMTHash2(L, R):
    """
    Computes the hash of an internal node using the hashes of its left and right children.
    """
    return Poseidon([int(L), int(R)])

# SMTLevIns
def SMTLevIns(n_levels, siblings, enabled):
    """
    Finds the level where oldInsert should be performed in a Sparse Merkle Tree.
    
    Operation Rules:
    1. levIns[i] = 1 when:
      - Current level (i) and all child levels have siblings of 0, and
      - The parent level has a sibling != 0
    2. The root level is always assumed to have a parent with a sibling != 0.

    Example (4-level tree):
    Level   siblings   levIns   done   Description
      0        0         0       1     Root level
      1        v         0       1     middle level
      2        0         1       1     Insertion level (parent sibling != 0)
      3        0         0       0     Lowest level (siblings have to be 0)
    
    Note: Exactly one level will have levIns = 1, ensuring efficient tree updates.
    """
    lev_ins = [0] * n_levels
    done = [0] * (n_levels - 1)

    # Check last level
    if enabled:
        assert siblings[-1] == 0, "Last sibling must be zero if enabled"

    # Calculate from highest to lowest level
    lev_ins[n_levels-1] = int((siblings[n_levels-2] != 0))
    done[n_levels-2] = lev_ins[n_levels-1]

    for i in range(n_levels - 2, 0, -1):
        lev_ins[i] = (1 - done[i]) * int((siblings[i-1] != 0))
        done[i-1] = lev_ins[i] + done[i]

    lev_ins[0] = 1 - done[0]

    print(f"lev_ins: {lev_ins}")
    print(f"done: {done}")

    return lev_ins

# SMTProcessorSM
# non-linear constraints : 5
def SMTProcessorSM(xor, is0, levIns, fnc, prev_top, prev_old0, prev_bot, prev_new1, prev_na, prev_upd):
    """
    Implements the state machine for processing each level of the Sparse Merkle Tree.

    This function calculates the state variables for each level of the tree during
    an update operation. It determines how the update should propagate through the tree.

    Args:
        xor : XOR of old and new key bits at this level.
        is0 : Flag indicating if the old value is zero.
        levIns : Flag indicating if insertion is needed at this level.
        fnc : Function flags [insert, update].
        prev_* : Previous state variables from the level above.

    Returns:
        st_*: New state variables for the current level.
    """
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
# non-linear constraints : 490
def SMTProcessorLevel(st_top, st_old0, st_bot, st_new1, st_na, st_upd,
                      sibling, old1leaf, new1leaf, newlrbit, oldChild, newChild):
    """
    Processes a single level of the Sparse Merkle Tree during an operation.

    This function computes the old and new root hashes for the current level,
    taking into account the various possible states of the update operation.

    Args:
        st_* : State variables from SMTProcessorSM.
        sibling : The sibling node at this level.
        old1leaf : Hash of the old leaf node.
        new1leaf : Hash of the new leaf node.
        newlrbit : Bit indicating left (0) or right (1) child for the new node.
        oldChild : Old child node hash.
        newChild : New child node hash.

    Returns:
        oldRoot, newRoot : Old and new root hashes for this level.
    """

    aux = [0] * 4

    # Old side
    oldSwitcher_L, oldSwitcher_R = Switcher(newlrbit, oldChild, sibling) #newlrbit decide which one is left and right
    oldProofHash = SMTHash2(oldSwitcher_L, oldSwitcher_R)


    aux[0] = old1leaf * (st_bot + st_new1 + st_upd)
    oldRoot = aux[0] + oldProofHash * st_top

    # New side
    aux[1] = newChild * (st_top + st_bot)
    newSwitcher_L = aux[1] + new1leaf * st_new1

    aux[2] = sibling * st_top
    newSwitcher_R = aux[2] + old1leaf * st_new1

    newSwitcher_outL, newSwitcher_outR = Switcher(newlrbit, newSwitcher_L, newSwitcher_R)
    newProofHash = SMTHash2(newSwitcher_outL, newSwitcher_outR)

    aux[3] = newProofHash * (st_top + st_bot + st_new1)
    newRoot = aux[3] + new1leaf * (st_old0 + st_upd)

    return {'oldRoot': oldRoot, 'newRoot': newRoot}


# SMTProcessor
def SMTProcessor(nLevels, oldRoot, siblings, oldKey, oldValue, isOld0, newKey, newValue, fnc):
    """
    Main function for processing updates in a Sparse Merkle Tree.

    This function orchestrates the entire process of updating the SMT, including
    insertion, update, and deletion operations. It computes the new root hash
    after applying the specified operation.

    Args:
        nLevels : Number of levels in the tree.
        oldRoot : Current root hash of the tree.
        siblings : Sibling nodes along the path of the update.
        oldKey : Key of the node being updated/deleted.
        oldValue : Current value of the node being updated/deleted.
        isOld0 : Flag indicating if the old value is zero (for deletions).
        newKey : Key of the node being inserted/updated.
        newValue : New value to be inserted/updated.
        fnc : Function flags [insert, update].

    Returns:
        newRoot: New root hash of the tree after the operation.
    """
    print("fnc: ", fnc)
    print("oldKey: ", oldKey)
    print("newKey: ", newKey)
    print("oldValue: ", oldValue)
    print("newValue: ", newValue)
    print("isOld0: ", isOld0)
    print("oldRoot: ", oldRoot)
    
    #Constraints: 2553 + 499*(nLevels-2)
    print("SMTProcessor Non-linear Constraints: ", 2553 + 499*(nLevels-2))
    print("Siblings: ", siblings)

    enabled = fnc[0] + fnc[1] - fnc[0] * fnc[1]

    hash1Old = SMTHash1(oldKey, oldValue)
    hash1New = SMTHash1(newKey, newValue)

    n2bOld = Num2Bits_strict(oldKey)
    n2bNew = Num2Bits_strict(newKey)

    smtLevIns = SMTLevIns(nLevels, siblings, enabled)

    # if oldkey and newkey are same, xor is 0, else 1 (repeat nLevels)
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

    # Compute final root hash
    topSwitcher_L, topSwitcher_R = Switcher(fnc[0] * fnc[1],levels[0]['oldRoot'], levels[0]['newRoot'])

    # Verify old root
    ForceEqualIfEnabled(enabled, oldRoot, topSwitcher_L)

    # Calculate new root
    newRoot = enabled * (topSwitcher_R - oldRoot) + oldRoot

    # Check keys are equal if updating
    areKeyEquals = int(oldKey == newKey)
    keysOk = MultiAND([1 - fnc[0], fnc[1], 1 - areKeyEquals])

    assert keysOk == 0

    return newRoot

###################################################
# Example usage
###################################################
fnc = [1, 0]  # Insert
old_root = 0  # Initial root hash
siblings = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0]  # Example with 2 levels
nlevels = len(siblings)
old_key = 0
old_value = 0
is_old0 = 1
new_key = 111
new_value = 222

#new_root should be 9308772482099879945566979599408036177864352098141198065063141880905857869998
new_root = SMTProcessor(nlevels, old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc)
print(f"New root after insertion: {new_root}")

# # Update
old_root = new_root
old_value = new_value
old_key = new_key
is_old0 = 0
new_value = 20
siblings = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0]
fnc = [0, 1]  # Update

new_root = SMTProcessor(nlevels, old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc)
print(f"New root after update: {new_root}")

fnc = [1, 0]  # Insert
old_root = new_root  # Initial root hash
siblings = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0]  # Example with 2 levels
old_key = 111
old_value = 20
is_old0 = 0
new_key = 110
new_value = 333

new_root = SMTProcessor(nlevels, old_root, siblings, old_key, old_value, is_old0, new_key, new_value, fnc)
print(f"New root after update: {new_root}")

# # Delete
# old_root = new_root
# new_value = 0
# fnc = [1, 1]  # Delete

# new_root = SMTProcessor(nlevels, old_root, siblings, old_key, old_value, False, new_key, new_value, fnc)
# print(f"New root after deletion: {new_root}")

# z = SMTHash2(0, 0)
# print(z)