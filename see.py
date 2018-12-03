

lines = open('a.txt').read().splitlines()

decimal,bytesval = [str.split(i,',') for i in lines]
mymap = {}
for i in range(len(decimal)):
    print i,decimal[i]
    mymap[bytesval[i]] = decimal[i]
base32 = "0123456789bcdefghjkmnpqrstuvwxyz"
print [mymap[i] for i in base32]


print bytesval

print str.split(lines[0],'\t')
