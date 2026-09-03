'use client'

import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {SelectorFor} from "@/app/components/selector";
import {ProjectsSelector, ReadWriteSelector} from "@/app/components/projectClient";
import {UserSelector} from "@/app/components/userClient";
import * as React from "react";
import {useContext, useEffect, useState} from "react";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {RemoveButton} from "@/app/components/formSubcomponents/commonClient";

export function UnmarshalAcl(input: any): ACL {
    if (typeof input !== 'object') {
        throw 'Input is not an object! Input is ' + typeof input
    }
    let bp: boolean | undefined = undefined
    if ('blanketPerm' in input && (!(input.blanketPerm === undefined || input.blanketPerm === null))) {
        if (typeof input.blanketPerm !== 'boolean'){
            throw 'blanketPerm must be boolean or missing'
        }
        bp = input.blanketPerm
    }

    const users = UnmarshalAclMapField(input, 'users')
    const projects = UnmarshalAclMapField(input, 'projects')
    return {
        blanketPerm: bp,
        users: users,
        projects: projects
    }
}

export function UnmarshalAclMapField(input: any, fieldName:string): Map<string,boolean> {
    if (fieldName in input && (input[fieldName] !== undefined)) {
        const field = input[fieldName]
        if (typeof field !== 'object' || field === null) {
            throw 'users field must be an object'
        }
        return new Map<string, boolean>(Object.entries(field as Record<string, boolean>))
    } else {
        return new Map<string, boolean>()
    }
}

export function NewAllCanWriteAcl():ACL{
    return {
        users: new Map<string, boolean>(),
        projects: new Map<string, boolean>(),
        blanketPerm: true,
    }
}

export function MarshalAcl(acl: ACL): any {
    if (acl === undefined) {
        return undefined
    }
    const out: any = {
        blanketPerm: acl.blanketPerm
    }

    if (("users" in acl) && acl.users !== undefined && acl.users !== null && acl.users.size !== 0) {
        if (acl.users instanceof Map) {
            out.users = Object.fromEntries(acl.users)
        } else {
            out.users = acl.users
        }
    }
    if (("projects" in acl) && acl.projects !== undefined && acl.projects !== null && acl.projects.size !== 0) {
        if (acl.projects instanceof Map) {
            out.projects = Object.fromEntries(acl.projects)
        } else {
            out.projects = acl.projects
        }

    }
    return out
}

export function ProjectsDisplay({readonly, initial, onClick, updateParent, allowAddingCompletedProjects}: {
    readonly: boolean,
    initial: Map<string, boolean>,
    onClick?: (proj: string) => void
    updateParent?: (p: Map<string, boolean>) => void
    allowAddingCompletedProjects?: boolean // TODO: USE THIS IN CALLERS!
}) {
    const projectNameAreaFor = (proj: [string, boolean]) => {
        return <text onClick={() => {
            onClick && onClick(proj[0])
        }}>{proj[0]}</text>
    }
    if (readonly) {
        if (initial===undefined || initial.size === 0) {
            return null
        }
        return <>{
            [...initial.entries()].map(values => {
                return <div key={values[0]}>
                    {projectNameAreaFor(values)}
                    <text>{"Can " + (values[1] ? "edit" : "view")}</text>
                </div>
            })
        }</>
    }
    const [current, setCurrent] = useState<Map<string, boolean>>(initial);
    useEffect(()=>{
        setCurrent(initial) // TODO: ensure new Map() not needed! TODO: can we remove useEffect here?
    },[initial])

    const update = (newPs: Map<string, boolean>) => {
        // newPs must be a newly-created map!
        setCurrent(newPs)
        updateParent && updateParent(newPs)
    }
    const addNewProject = (projName: string) => {
        if (initial.size === 0) {
            update(new Map<string, boolean>().set(projName, false))
        } else {
            update(new Map<string, boolean>([...current.entries()]).set(projName, false))
        }
        return
    }
    const removeProject = (projName: string) => {
        const newProjs = new Map(current)
        newProjs.delete(projName)
        update(newProjs)
        return
    }
    return <div>
        <TestAndValidate>
            {current.size >0 && [...current.entries()].map(values => {
                const projName = values[0]
                const canWrite = values[1]
                return <div key={projName}>
                    {projectNameAreaFor(values)}
                    <SelectorFor options={["can view", "can edit"]} initial={canWrite ? "can edit" : "can view"}
                                 updateParent={s => {
                                     const newProjs = new Map(current).set(projName, s === "can edit") // TODO: do this for users too!
                                     update(newProjs)
                                 }} disabled={false}/>
                    <text>{"Can " + (canWrite? "edit" : "view")}</text>
                    <RemoveButton txt={"Remove project"} click={() => {
                        removeProject(projName)
                    }}/>
                </div>
            })}
            {"Add a project:"}
            <ProjectsSelector onSelect={addNewProject} complete={allowAddingCompletedProjects}
                              blacklist={(initial.size !== 0) ? [...initial.entries()].map(val => {
                                  return val[0]
                              }) : []}/>
        </TestAndValidate>
    </div>
}

export function AclProjectsDisplay({readonly, initial, onUsernClick, updateParent}: {
    readonly: boolean,
    initial: Map<string,boolean>,
    onUsernClick?: (proj: string) => void
    updateParent?: (newProjects: Map<string,boolean>) => void
}) {
    return <ProjectsDisplay readonly={readonly}
                            allowAddingCompletedProjects={false/* TODO: ok?*/}
                            initial={initial}
                            onClick={onUsernClick} updateParent={projPerms => {
        updateParent && updateParent(projPerms) // TODO: ensure clone not needed
    }}/>
}

export function AclUsersDisplayInternal({readonly, initial, onClick, updateParent, blanket}: {
    readonly: boolean,
    initial: Map<string, boolean>,
    onClick?: (usr: string) => void
    updateParent?: (p: Map<string, boolean>) => void
    blanket?: boolean
}) {
    const userNameAreaFor = (val: [string, boolean]) => {
        const usern = val[0]
        return <text onClick={(e) => {
            e.stopPropagation()
            onClick && onClick(usern)
        }}>{usern}</text>
    }
    if (readonly) {
        // TODO: Should have incremented depth?
        {
            initial !== undefined && initial.size > 0 && [...initial.entries()].map(values => {
                return <div key={values[0]}>
                    {userNameAreaFor(values)}
                    <text>{"Can " + (values[1] ? "edit" : "view")}</text>
                </div>
            })
        }
    }
    const [users, setUsers] = useState<Map<string, boolean>>(initial.size >0 ? new Map(initial) : new Map())
    useEffect(() => {
        setUsers(initial.size >0 ? new Map(initial) : new Map())
    }, [initial]);
    const update = (updated: Map<string, boolean>) => {
        setUsers(updated)
        updateParent && updateParent(updated)
    }
    const addNewUser = (uName: string) => {
        const defaultPerm = blanket || false
        if (users===undefined || users.size === 0) {
            update(new Map<string, boolean>().set(uName, defaultPerm))
        } else {
            update(new Map<string, boolean>(users).set(uName, defaultPerm))
        }
        return
    }
    const removeUsr = (projName: string) => {
        const newUsrs = new Map(users)
        newUsrs.delete(projName)
        update(newUsrs)
        return
    }

    return <DepthProvider>
        <div>
            {(users.size > 0) && [...users.entries()].map(values => {
                const email = values[0]
                const canWrite = values[1]
                return <div key={email}>
                    {userNameAreaFor(values)}
                    <ReadWriteSelector value={canWrite} readonly={false} onUpdate={b => {
                        const newUsrs = new Map(users).set(email, b)
                        update(newUsrs)
                    }}/>
                    {/*<text>{"Can " + (canWrite ? "edit" : "view")}</text>TODO: consider re-adding*/}
                    <RemoveButton txt={"Remove user"} click={() => {
                        removeUsr(email)
                    }}/>
                </div>
            })}
            {"Add a user:"}
            <UserSelector onSelect={(u) => {
                addNewUser(u._id)
            }} blacklist={(users.size > 0) ? [...users.entries()].map(val => val[0]) : []}/>
        </div>
    </DepthProvider>
}

export function AclBlanketDisplay(inp: {
    readonly: boolean,
    ACL: ACL,
    updateParent: (a?: boolean) => void
}) {
    const [val, setVal] = useState(inp.ACL)
    useEffect(() => {
        setVal(inp.ACL)
    }, [inp.ACL])
    const update = (b?: boolean) => {
        if (val !== undefined) {
            const temp = structuredClone(val)
            temp.blanketPerm = b
            setVal(temp)
        } else {
            setVal({
                users: new Map<string, boolean>(),
                projects: new Map<string, boolean>(),
                blanketPerm: b
            })
        }
        inp.updateParent && inp.updateParent(b)
    }
    if (inp.readonly) {
        return <div>{(val === undefined || val.blanketPerm === true) ? "Publicly Editable" : ((val.blanketPerm === false) ? "Publicly Viewable" : "Private")}</div>
    }
    const permToStr = (a: ACL) => {
        return a.blanketPerm === undefined ? "Private" : (a.blanketPerm ? "Publicly Editable" : "Publicly Viewable")
    }
    const strToPerm = (s: string) => {
        switch (s) {
            case "Publicly Editable":
                return true
            case "Publicly Viewable":
                return false
            default:
                return undefined
        }
    }
    return <SelectorFor initial={permToStr(inp.ACL)} options={["Private", "Publicly Viewable", "Publicly Editable"]}
                        updateParent={permStr => {
                            update(strToPerm(permStr))
                        }} disabled={false}/>
}

export function TogglableAreaWithDepth(props: React.PropsWithChildren<{
    startOpen: boolean,
    openTxt?: string,
    closeTxt?: string
}>) {
    const [open, setOpen] = useState(props.startOpen);
    const toggle = () => {
        setOpen(!open);
    }
    if (!open) {
        return <button className={"basicButton"} onClick={toggle}>{props.openTxt || "open"}</button>
    }
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div className={"subForm depth" + depth}>
            {props.children}
            <button className={"basicButton"} onClick={toggle}>{props.closeTxt || "close"}</button>
        </div>
    </DepthProvider>
}

export function AclDefaultAclDisplay(inp: {
    readonly: boolean,
    ACL: ACL,
    defaultACL: ACL,
    updateAcl: (acl: ACL) => void
    updateDefaultAcl: (acl: ACL) => void
}) {
    const depth = useContext(DepthContext)
    const [open, setOpen] = useState<string | undefined>(undefined)
    const hideButton = <button className={"basicButtonSmall"}>{"Hide ACL"}</button>
    const aclArea = () => {
        if (!open) {
            return null
        }
        if (open === "ACL") {
            return <><AclDisplay initial={inp.ACL} updateParent={inp.updateAcl} readonly={inp.readonly}/>
                {hideButton}
            </>
        }
        return <><AclDisplay initial={inp.defaultACL} updateParent={inp.updateDefaultAcl} readonly={inp.readonly}/>
            {hideButton}
        </>
    }
    const btnClicked = (v: string) => {
        setOpen((open && v === open) ? undefined : v)
    }
    const aButton = <div className={"depth" + (open === "ACL" ? (depth) + " closeable" : (depth + 1) + " activatable")}
                         onClick={() => {
                             btnClicked("ACL")
                         }}>{"ACL"}</div>
    const bButton = <div
        className={"depth" + (open === "defaultACL" ? (depth) + " closeable" : (depth + 1) + " activatable")}
        onClick={() => {
            btnClicked("defaultACL")
        }}>{"Default ACL"}</div>
    return <div className={"subForm depth" + depth}>
        <div className={"aclDualButton"}>
            {aButton}
            {bButton}
        </div>
        {aclArea()}
    </div>
}

export function AclDisplay(inp: {
    readonly: boolean,
    initial: ACL,
    updateParent: (acl: ACL) => void
    // TODO: onUserClick???
}) {
    const [current, setCurrent] = React.useState(inp.initial)
    const mapFor = (inpMap?: Map<string, boolean>):Map<string, boolean> => {
        if (inpMap === undefined || inpMap === null || inpMap.size === 0) {
            return new Map<string, boolean>()
        }
        return new Map<string, boolean>(inpMap)
    }
    useEffect(()=>{
        setCurrent({
            blanketPerm: inp.initial.blanketPerm,
            users: mapFor(inp.initial.users),
            projects: mapFor(inp.initial.projects),
        })
    },[inp.initial])
    const cloneCurrent = ()=>{
        return {
            blanketPerm: current.blanketPerm,
            users: mapFor(current.users),
            projects: mapFor(current.projects),
        }
        //return structuredClone(current) // TODO: ensure ok
    }
    const update = (updated: ACL)=>{
        inp.updateParent(updated)
        setCurrent(updated)
    }
    const updateProjects = (updated: Map<string,boolean>)=>{
        //const newProjects = {...cloneCurrent(), projects: updated}
        // console.log("new projects: ") // TODO: del
        // console.table(Array.from(newProjects.projects.entries()), ["name", "canWrite"]); // TODO: del
        update({...cloneCurrent(), projects: updated})
    }
    const updateUsers = (updated: Map<string,boolean>)=>{
        update({...cloneCurrent(), users: updated})
    }
    const updateBlanket = (updated?: boolean)=>{
        if (updated === undefined) {
            update({
                users: mapFor(current.users),
                projects: mapFor(current.projects),
            })
        } else {
            update({...cloneCurrent(), blanketPerm: updated})
        }

    }

    const depth = useContext(DepthContext)
    if (inp.readonly) {
        return <div>{/* TODO: TURN INTO A TABLE?!!!!*/}
            <TestAndValidate todos={["if users or projects do not exist on existing ACLs, problems are caused..."]}>
                <AclBlanketDisplay readonly={true} ACL={inp.initial} updateParent={()=>{}}/>
                <AclUsersDisplayInternal readonly={true} initial={inp.initial.users || new Map<string, boolean>()}/>
                <AclProjectsDisplay readonly={true} initial={inp.initial.projects || new Map()} onUsernClick={() => {/* TODO: onClick?*/
                }} updateParent={()=>{}}/>
            </TestAndValidate>
        </div>

    }
    // TODO: COMPLETELY OVERHAUL
    return <div className={"subForm depth" + depth}>
        <AclBlanketDisplay readonly={false} ACL={current} updateParent={updateBlanket}/>
        <AclUsersDisplayInternal readonly={false} blanket={inp.initial.blanketPerm}
                                 initial={inp.initial.users || new Map<string, boolean>()}
                                 onClick={() => {
                                     // TODO: this
                                 }} updateParent={updateUsers}/>
        <AclProjectsDisplay readonly={false} initial={inp.initial.projects || new Map()} onUsernClick={usern => {
            // TODO: this?
        }} updateParent={updateProjects}/>
    </div>
}

// export function DefaultAclDisplay({readonly, initial, updateParent}: {
//     readonly: boolean,
//     initial: ACL,
//     updateParent: (acl: ACL) => void
// }) {
//     return <div>
//         {"Default Access Control List"}
//         <AclDisplay readonly={readonly} initial={initial} updateParent={updateParent}/>
//     </div>
// }