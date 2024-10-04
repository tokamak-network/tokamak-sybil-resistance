import os
import json
from smt import SMTProcessor

def load_test_data(file_name):
    current_dir = os.path.dirname(os.path.abspath(__file__))
    file_path = os.path.join(current_dir, file_name)
    with open(file_path, 'r') as f:
        return json.load(f)

def run_tests(test_data):
    for i, test_case in enumerate(test_data):
        print(f"Running test case {i + 1}")
        
        # Convert input data to appropriate types
        nLevels = test_case['nlevels']
        oldRoot = int(test_case['oldRoot'])
        siblings = [int(s) for s in test_case['siblings']]
        oldKey = int(test_case['oldKey'])
        oldValue = int(test_case['oldValue'])
        isOld0 = int(test_case['isOld0'])
        newKey = int(test_case['newKey'])
        newValue = int(test_case['newValue'])
        fnc = [int(f) for f in test_case['fnc']]
        expected_newRoot = int(test_case['newRoot'])

        # Run SMTProcessor
        newRoot = SMTProcessor(nLevels, oldRoot, siblings, oldKey, oldValue, isOld0, newKey, newValue, fnc)

        print(f"newRoot: {newRoot}")
        print(f"expected_newRoot: {expected_newRoot}")
        # Check result
        if newRoot == expected_newRoot:
            print(f"Test case {i + 1} passed")
        else:
            print(f"Test case {i + 1} failed")
            print(f"Expected: {expected_newRoot}")
            print(f"Got: {newRoot}")
        
        print("---")

if __name__ == "__main__":
    test_data = load_test_data('test_data.json') # testcases from https://github.com/iden3/circomlib
    run_tests(test_data)