export interface ACL {
    users?: Map<string,boolean>
    projects?: Map<string,boolean>
    blanketPerm?: boolean
}

export function TestAcl(){
    return {blanketPerm: true}
}