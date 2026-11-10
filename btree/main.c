#include <stdio.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#include <math.h>
#include <string.h>

#define PAGESIZE 16384
#define NUMPOINTERS 5

typedef struct BTreeNode {
    bool isLeaf;
    uint64_t keys[NUMPOINTERS - 1];
    uint64_t pointers[NUMPOINTERS];
    int numPointers;
    struct BTreeNode* parent;
} BTN;

typedef struct BTreeRecord {
    uint64_t index;
    char* name;
    uint32_t age;
} BTR;

void insert(BTN* node, BTR* record);
void splitPageInternal(BTN* node, BTN* ptr, uint64_t key);
void printTree(BTN* node, int depth);

BTR* getRecord(uint64_t index, const char* name, uint32_t age) {
    BTR* rec = calloc(1, sizeof(BTR));
    if (rec == NULL) {
        return NULL;
    }

    rec->index = index;
    rec->age = age;
    rec->name = malloc(strlen(name) + 1);
    if (rec->name == NULL) {
        free(rec);
        return NULL;
    }
    strcpy(rec->name, name);
    return rec;
}

void freeRecord(BTR* rec) {
    if (rec == NULL) {
        return;
    }

    free(rec->name);
    free(rec);
}

BTN* getLeaf(BTN* node, uint64_t k) {
    if (node->isLeaf == true) {
        return node;
    }

    uint64_t mn = 0;
    int i = 0;
    for (; i < NUMPOINTERS - 1; i++) {
        uint64_t key = node->keys[i];
        if (key == 0) {
            key = UINT64_MAX;
        }

        BTN* child = (BTN*) node->pointers[i];
        if (child == NULL) {
            return NULL;
        }
        else if (mn <= k && k < key) {
            BTN* ret = getLeaf(child, k);
            return ret;
        }
        mn = node->keys[i];
    }

    
    BTN* child = (BTN*) node->pointers[i];
    if (child == NULL) {
        return NULL;
    }

    BTN* ret = getLeaf(child, k);
    return ret;
}

BTR* search(BTN* node, uint64_t k) {
    if (!node->isLeaf) {
        node = getLeaf(node, k);
    }
    
    if (node == NULL) {
        return NULL;
    }

    for (int i = 0; i < NUMPOINTERS; i++) {
        BTR* record = (BTR*) node->pointers[i];
        if (record != NULL && record->index == k) {
            return record;
        }
    }

    return NULL;
}

void insertInternal(BTN* node, BTN* ptr, uint64_t key) {
    if (node->numPointers == NUMPOINTERS) {
        splitPageInternal(node, ptr, key);
        return;
    }

    if (node->numPointers == 0) {
        node->pointers[0] = (uint64_t) ptr;
        node->keys[0] = key;
        node->numPointers++;
        ptr->parent = node;
        return;
    }

    uint64_t mn = node->keys[0];
    for (int i = 0; i < NUMPOINTERS - 1; i++) {
        if (key < mn) {
            memmove(&(node->keys[i + 1]), &(node->keys[i]), sizeof(uint64_t) * (node->numPointers - i - 1));
            memmove(&(node->pointers[i + 2]), &(node->pointers[i + 1]), sizeof(uint64_t) * (node->numPointers - i - 1));
            node->keys[i] = key;
            node->pointers[i + 1] = (uint64_t) ptr;
            node->numPointers += 1;
            ptr->parent = node;
            return;
        }
        if (i < NUMPOINTERS - 2 && node->keys[i + 1] != 0) {
            mn = node->keys[i + 1];
        }
        else {
            mn = UINT64_MAX;
        }
    }
    
    node->keys[NUMPOINTERS - 2] = key;
    node->pointers[NUMPOINTERS - 1] = (uint64_t) ptr;
    node->numPointers++;
    ptr->parent = node;
    return;
}

void splitPageInternal(BTN* node, BTN* ptr, uint64_t key) {
    BTN* newNode = calloc(1, sizeof(BTN));
    newNode->isLeaf = false;
    newNode->parent = node->parent;
    newNode->numPointers = 0;

    int start = NUMPOINTERS / 2;
    uint64_t k = node->keys[NUMPOINTERS / 2 - 1];
    if (node->keys[NUMPOINTERS / 2 - 1] < key && key < node->keys[start]) {
            k = key;
            start++;
    }

    for (int i = start; i < NUMPOINTERS; i++) {
        newNode->pointers[i - (NUMPOINTERS / 2)] = node->pointers[i];

        if (i < NUMPOINTERS - 1) {
            newNode->keys[i - (NUMPOINTERS / 2)] = node->keys[i];
            node->keys[i] = 0;
        }

        if (i == start && key == k) {
            newNode->keys[0] = node->keys[i - 1];
            node->keys[i - 1] = 0;
        }

        ((BTN*) node->pointers[i])->parent = newNode;
        node->numPointers--;
        newNode->numPointers++;
        node->pointers[i] = 0;
    }

    if (key < k) {
        insertInternal(node, ptr, key);
    }
    else if (key == k) {
        newNode->pointers[0] = (uint64_t) ptr;
        newNode->numPointers++;
        ptr->parent = newNode;
    }
    else {
        insertInternal(newNode, ptr, key);
    }

    if (node->parent == NULL) {
        BTN* root = calloc(1, sizeof(BTN));
        root->isLeaf = false;
        root->numPointers = 0;
        root->parent = NULL;

        node->parent = root;
        newNode->parent = root;
        root->pointers[0] = (uint64_t) node;
        root->pointers[1] = (uint64_t) newNode;
        root->keys[0] = k;
        root->numPointers = 2;
        return;
    }
    else {
        newNode->parent = node->parent;
        insertInternal(node->parent, newNode, k);
    }
}

void splitPageLeaf(BTN* node, BTR* record) {
    BTN* newNode = calloc(1, sizeof(BTN));
    newNode->isLeaf = true;
    newNode->parent = node->parent;
    newNode->numPointers = 0;
    uint64_t k = ((BTR*) node->pointers[NUMPOINTERS / 2])->index;
    for (int i = NUMPOINTERS / 2; i < NUMPOINTERS; i++) {
        newNode->pointers[i - (NUMPOINTERS / 2)] = node->pointers[i];
        newNode->numPointers++;
        node->numPointers--;
        node->pointers[i] = 0;
    }

    if (record->index < k) {
        insert(node, record);
    }
    else {
        insert(newNode, record);
    }

    if (node->parent != NULL) {
        insertInternal(node->parent, newNode, k);
    }
    else {
        BTN* root = calloc(1, sizeof(BTN));
        root->keys[0] = k;
        root->pointers[0] = (uint64_t) node;
        root->pointers[1] = (uint64_t) newNode;
        root->numPointers = 2;
        
        root->isLeaf = false;
        root->parent = NULL;
        node->parent = root;
        newNode->parent = root;
    }
}

void insert(BTN* node, BTR* record) {
    BTN* leaf = getLeaf(node, record->index);
    if (leaf == NULL) {
        return;
    }

    if (leaf->numPointers == 0) {
        leaf->pointers[0] = (uint64_t) record;
        leaf->numPointers++;
        return;
    }


    
    int l = 0;
    int r = leaf->numPointers - 1;

    while (l <= r) {
        int mid = (l + r) / 2;
        BTR* currRec = (BTR*)(leaf->pointers[mid]);

        if (currRec->index == record->index) {
            if (record == currRec) {
                return;
            }
            freeRecord(currRec);
            leaf->pointers[mid] = (uint64_t) record;
            return;
        }
        else if (record->index < currRec->index) {
            r = mid - 1;
        }
        else {
            l = mid + 1;
        }
    }

    if (leaf->numPointers == NUMPOINTERS) {
        splitPageLeaf(leaf, record);
        return;
    }

    uint64_t mn = ((BTR*) leaf->pointers[0])->index;
    for (int i = 0; i < NUMPOINTERS - 1; i++) {
        if (record->index <= mn && leaf->pointers[i] != 0) {
            memmove(&(leaf->pointers[i + 1]), &(leaf->pointers[i]), sizeof(uint64_t) * (leaf->numPointers - i));
            leaf->pointers[i] = (uint64_t) record;
            leaf->numPointers += 1;
            return;
        }
        else if (record->index < mn) {
            leaf->numPointers += 1;
            leaf->pointers[i] = (uint64_t) record;
            return;
        }

        if (leaf->pointers[i + 1] == 0) {
            mn = UINT64_MAX;
        }
        else {
            mn = ((BTR*) leaf->pointers[i + 1])->index;
        }
    }

    leaf->pointers[NUMPOINTERS - 1] = (uint64_t) record;
    leaf->numPointers += 1;
    return;
}

void printTree(BTN* node, int depth) {
    for (int j = 0; j < depth; j++) {
                printf("\t");
    }
    printf("%llu\n", (uint64_t) node);

    if (node->isLeaf == false) {
        for (int i = 0; i < NUMPOINTERS; i++) {
            if (node->pointers[i] == 0) {
                break;
            }

            for (int j = 0; j < depth; j++) {
                printf("\t");
            }
            printf("%llu ", node->pointers[i]);
            if (i < NUMPOINTERS - 1) {
                printf("%llu ", node->keys[i]);
            }
            
            if (node->pointers[i] != 0) {
                printf("\n");
                printTree((BTN*) node->pointers[i], depth + 1);
            }
        }
    }
    else {
        for (int i = 0; i < NUMPOINTERS; i++) {
            for (int j = 0; j < depth; j++) {
                printf("\t");
            }
            if (node->pointers[i] != 0) {
                printf("%llu ", ((BTR*)(node->pointers[i]))->index);
            }
        }
        printf("\n");
    }

    printf("\n");
}
bool deleteInternal(BTN* node, BTN* child, uint64_t val) {
    if (node == NULL) {
        free(child);
        return true;
    }
    
    int mid = 0;
    for (; mid < node->numPointers; mid++) {
        BTN* curr = (BTN*) (node->pointers[mid]);
        if (curr == child) {
            break;
        }
    }

    free(child);
    node->numPointers--;
    node->pointers[mid] = 0;

    if (node->numPointers == 0) {
        if (node->parent != NULL) {
            return deleteInternal(node->parent, node, node->keys[mid]);
        }
        else {
            node->isLeaf = true;
            return true;
        }
    }
    else if (node->numPointers < NUMPOINTERS) {
        memmove(&(node->pointers[mid]), &(node->pointers[mid + 1]), sizeof(uint64_t) * (node->numPointers - (mid)));
        node->pointers[node->numPointers] = 0;
        
        if (mid < node->numPointers) {
            memmove(&(node->keys[mid]), &(node->keys[mid + 1]), sizeof(uint64_t) * (node->numPointers - mid - 1));
            node->keys[node->numPointers - 1] = 0;
        }
        else {
            node->keys[mid - 1] = 0;
        }
    }
    else {
        node->keys[mid - 1] = 0;
    }
    return true;
}

bool delete(BTN* root, uint64_t val) {
    BTN* leaf = getLeaf(root, val);
    if (leaf == NULL) {
        return false;
    }

    int l = 0;
    int r = leaf->numPointers - 1;
    int mid;
    bool deleted = false;

    while (l <= r) {
        mid = (l + r) / 2;
        BTR* rec = ((BTR*)(leaf->pointers[mid]));

        if (rec->index == val) {
            freeRecord(rec);
            leaf->pointers[mid] = 0;
            leaf->numPointers--;
            deleted = true;
            break;
        }
        else if (val < rec->index) {
            r = mid - 1;
        }
        else {
            l = mid + 1;
        }
    }

    if (deleted == false) {
        return deleted;
    }

    if (mid == NUMPOINTERS - 1) {
        return true;
    }
    else if (leaf->numPointers == 0) {
        if (leaf->parent != NULL) {
            return deleteInternal(leaf->parent, leaf, val);
        }
        return true;
    }
    else {
        memmove(&(leaf->pointers[mid]), &(leaf->pointers[mid + 1]), sizeof(uint64_t) * (leaf->numPointers - (mid)));
        leaf->pointers[leaf->numPointers] = 0;
        return true;
    }
}

int main() {
    BTN* parent = calloc(1, sizeof(BTN));
    parent->isLeaf = true;

    for (int i = 1; i < 26; i++) {
        BTR* rec = getRecord(i, "john", 123);
        insert(parent, rec);
        if (parent->parent != NULL) {
            parent = parent->parent;
        }
    }
    printTree(parent, 0);

    for (int i = 13; i < 26; i++) {
        delete(parent, i);
    }
    printTree(parent, 0);

    BTR* rec = getRecord(14, "john", 123);
    insert(parent, rec);
    printTree(parent, 0);

    return 0;
}
