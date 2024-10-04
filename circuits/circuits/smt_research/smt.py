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

#Check Is Equal
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
    return Poseidon([int(key), int(value), 1])

def SMTHash2(L, R):
    return Poseidon([int(L), int(R)])

# SMTLevIns
def SMTLevIns(n_levels, siblings, enabled):
    """
    Determines the insertion level for a new node in the Sparse Merkle Tree.

    This function calculates which level of the tree requires a new node insertion.
    It works from the bottom of the tree upwards, finding the deepest level where
    insertion is needed while maintaining tree balance.

    The algorithm ensures that new nodes are inserted at the lowest possible level,
    which helps keep the tree balanced and minimizes the number of hash computations
    needed for updates.

    Args:
        n_levels (int): The number of levels in the Sparse Merkle Tree.
        siblings (list): The sibling nodes for each level of the path in the tree.
        enabled (int): A flag indicating whether the insertion is enabled (1) or not (0).

    Returns:
        list: An array 'lev_ins' where lev_ins[i] == 1 if insertion is needed at level i, else 0.
    """
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
    """
    Implements the state machine for processing each level of the Sparse Merkle Tree.

    This function calculates the state variables for each level of the tree during
    an update operation. It determines how the update should propagate through the tree.

    Args:
        xor (int): XOR of old and new key bits at this level.
        is0 (int): Flag indicating if the old value is zero.
        levIns (int): Flag indicating if insertion is needed at this level.
        fnc (list): Function flags [insert, update].
        prev_* (int): Previous state variables from the level above.

    Returns:
        dict: New state variables for the current level.
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
def SMTProcessorLevel(st_top, st_old0, st_bot, st_new1, st_na, st_upd,
                      sibling, old1leaf, new1leaf, newlrbit, oldChild, newChild):
    """
    Processes a single level of the Sparse Merkle Tree during an update operation.

    This function computes the old and new root hashes for the current level,
    taking into account the various possible states of the update operation.

    Args:
        st_* (int): State variables from SMTProcessorSM.
        sibling (int): The sibling node at this level.
        old1leaf (int): Hash of the old leaf node.
        new1leaf (int): Hash of the new leaf node.
        newlrbit (int): Bit indicating left (0) or right (1) child for the new node.
        oldChild (int): Old child node hash.
        newChild (int): New child node hash.

    Returns:
        dict: Old and new root hashes for this level.
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
        nLevels (int): Number of levels in the tree.
        oldRoot (int): Current root hash of the tree.
        siblings (list): Sibling nodes along the path of the update.
        oldKey (int): Key of the node being updated/deleted.
        oldValue (int): Current value of the node being updated/deleted.
        isOld0 (int): Flag indicating if the old value is zero (for deletions).
        newKey (int): Key of the node being inserted/updated.
        newValue (int): New value to be inserted/updated.
        fnc (list): Function flags [insert, update].

    Returns:
        int: New root hash of the tree after the operation.
    """
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

    # Compute final root hash
    topSwitcher_L, topSwitcher_R = Switcher(fnc[0] * fnc[1],levels[0]['oldRoot'], levels[0]['newRoot'])

    # Verify old root
    ForceEqualIfEnabled(enabled, oldRoot, topSwitcher_L)

    # Calculate new root
    newRoot = enabled * (topSwitcher_R - oldRoot) + oldRoot

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


'''
[SMTProcessor]
    |
    |-- inputs: nLevels, oldRoot, siblings, oldKey, oldValue, isOld0, newKey, newValue, fnc
    |
    |-- [Utility Functions]
    |   |-- Num2Bits_strict
    |   |-- XOR
    |   |-- IsEqual
    |   |-- AND / MultiAND
    |   |-- Switcher
    |   |-- ForceEqualIfEnabled
    |
    |-- [Hash Functions]
    |   |-- SMTHash1
    |   |-- SMTHash2
    |
    |-- [SMTLevIns]
    |   |-- calculation: lev_ins
    |
    |-- [SMTProcessorSM] (each level)
    |   |-- inputs: xor, is0, levIns, fnc, prev_states
    |   |-- outputs: st_top, st_old0, st_bot, st_new1, st_na, st_upd
    |
    |-- [SMTProcessorLevel] (each level(reverse))
    |   |-- inputs: states, sibling, old1leaf, new1leaf, newlrbit, oldChild, newChild
    |   |-- [Switcher] (oldSwitcher)
    |   |-- [SMTHash2] (oldProofHash)
    |   |-- [Switcher] (newSwitcher)
    |   |-- [SMTHash2] (newProofHash)
    |   |-- outputs: oldRoot, newRoot
    |
    |-- [final process]
    |   |-- [Switcher] (topSwitcher)
    |   |-- [ForceEqualIfEnabled]
    |   |-- newRoot calculation
    |   |-- [MultiAND] (keysOk check)
    |
    |-- output: newRoot
'''