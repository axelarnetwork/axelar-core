	if ! git diff --exit-code -s ./contract-artifacts; then
	    echo changed
    else
    	echo same
	fi
