'use client'

// TODO: grain batches list and display need fixing
// TODO: plugs list and display need fixing
// TODO: projects list and display need fixing
// TODO: sporeSwab view not working. https://mush.appli.ng/view/sporeSwab/2g1j95Bw5gB
// TODO: transfers list needs fixing. Display appears fine
// TODO: users list needs fixing. Display needs changing?

import {OptionalKey, OptionalSimpleKey} from "@/app/components/common";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {SelectorFor} from "@/app/components/selector";
import {ProjectsSelector, ReadWriteSelector} from "@/app/components/projectClient";
import {UserSelector} from "@/app/components/userClient";
import * as React from "react";
import {useContext, useEffect, useState} from "react";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {RemoveButton} from "@/app/components/formSubcomponents/commonClient";

export function AssertACL(input: any): asserts input is ACL { // TODO: FIX THIS!!!! NEEDS TO DO MAP STUFF PROPERLY!
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['blanketPerm', 'boolean'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {

        if (!OptionalSimpleKey(key, input, expType)) {
            //console.error("failed when validating NON maps!") // TODO: THIS!
            throw new Error('ACL assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    const complexOptionalMapKeys = new Map<string, (v: any) => boolean>([
        ['users', IsStringMapToBool], // TODO: UNSURE IF WORKING
        ['projects', IsStringMapToBool], // TODO: UNSURE IF WORKING
    ])
    for (const [key, validator] of complexOptionalMapKeys) {
        if (!OptionalKey(key, input, validator)) {
            console.error("failed when validating maps!") // TODO: THIS!
            throw new Error('ACL assertion failure: optional array key ' + key + ' was not valid: ' + JSON.stringify(input[key]));
        }
    }
    if (input.users !== null && input.users !== undefined) {
        input.users = new Map<string, boolean>(Object.entries(input.users) as [string, boolean][]) // TODO: UNSURE IF WORKING
    } else {
        input.users = new Map<string, boolean>(); // TODO: will this screw up null (full-write) ACLs?
    }
    if (input.projects !== null && input.projects !== undefined) {
        input.projects = new Map<string, boolean>(Object.entries(input.projects) as [string, boolean][]) // TODO: UNSURE IF WORKING
    } else {
        input.projects = new Map<string, boolean>(); // TODO: will this screw up null (full-write) ACLs?
    }


    return
}

export function MarshalAcl(acl: ACL): any {
    if (acl === undefined) {
        return undefined
    }
    const out: any = {
        blanketPerm: acl.blanketPerm
    }

    if (("users" in acl) && acl.users !== undefined && acl.users.size !== 0) {
        // TODO: why is this occasionally coming back as an object and not a map????? FIX
        if (acl.users instanceof Map) {
            out.users = Object.fromEntries(acl.users)
            console.log("It's a Map");
        } else {
            out.users = acl.users
        }
    }
    if (("projects" in acl) && acl.projects !== undefined && acl.projects.size !== 0) {
        // TODO: why is this occasionally coming back as an object and not a map????? FIX
        if (acl.projects instanceof Map) {
            out.projects = Object.fromEntries(acl.projects)
            console.log("It's a Map");
        } else {
            out.projects = acl.projects
        }

    }
    return out

}

export function IsValidAcl(input: any): boolean {
    try { // TODO: ensure ok
        AssertACL(input) // TODO: may not properly replace the ACL
        return true
    } catch (error) {
        console.error("acl invalid: "+JSON.stringify(error)) // TODO: del
        return false
    }
}

// TODO: validate ok
export function IsStringMapToString(data: any): data is Record<string, boolean | undefined> {
    // 1. Check if the input is an object and not null.
    if (typeof data !== 'object') {
        return false;
    }

    // 2. Iterate over all keys of the object.
    for (const key in data) {
        if (Object.prototype.hasOwnProperty.call(data, key)) {
            if (typeof data[key] !== 'string') {
                //console.log("typeof entry: " + typeof data[key]) // TODO: delete
                return false;
            }
        }
    }

    // 3. If all values are booleans, it is a map of string to bool.
    return true;
}

export function IsStringMapToBool(data: any): data is Record<string, boolean> {
    // 1. Check if the input is an object and not null.
    if (typeof data !== 'object' || data === null || data === undefined) {
        return false;
    }

    // 2. Iterate over all keys of the object.
    for (const key in data) {
        if (Object.prototype.hasOwnProperty.call(data, key)) {
            // In JavaScript, object keys are always strings.
            // We only need to check the type of each value.
            if (!(typeof data[key] === 'boolean')) {
                // If any value is not a boolean, it fails the check.
                return false;
            }
        }
    }

    // 3. If all values are booleans, it is a map of string to bool.
    return true;
}

export function ProjectsDisplay({readonly, initial, onClick, updateParent, allowAddingCompletedProjects}: { // TODO: validate working properly
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
        if (initial.size === 0) {
            return null
        }
        // TODO: FIX THIS!?
        return <>{
            [...initial.entries()].map(values => {
                return <div key={values[0]}>
                    {projectNameAreaFor(values)}
                    <text>{"Can " + (values[1] ? "edit" : "view")}</text>
                </div>
            })
        }</>
    }
    const [current, setCurrent] = useState<Map<string, boolean>>(initial); // TODO: ensure new Map() not needed
    useEffect(()=>{
        setCurrent(initial) // TODO: ensure new Map() not needed
    },[initial])

    const update = (newPs: Map<string, boolean>) => {
        setCurrent(newPs) // TODO: vs structuredClone(newPs)
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
            {[...current.entries()].map(values => { // TODO: THE ENTRIES MAP HERE IS THE CURRENT PROBLEM
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
    onUsernClick?: (proj: string) => void // TODO: do we need both onClick and updateParent?
    updateParent?: (newProjects: Map<string,boolean>) => void
}) {
    return <ProjectsDisplay readonly={readonly}
                            allowAddingCompletedProjects={false/* TODO: ok?*/}
                            initial={initial}
                            onClick={onUsernClick} updateParent={(projPerms) => {
        updateParent && updateParent(projPerms) // TODO: ensure clone not needed
    }}/>
}

// TODO: USE THIS! validate working properly!
export function AclUsersDisplayInternal({readonly, initial, onClick, updateParent, blanket}: {
    readonly: boolean,
    initial: Map<string, boolean>,
    onClick?: (usr: string) => void
    updateParent?: (p: Map<string, boolean>) => void
    blanket?: boolean
}) {
    const userNameAreaFor = (val: [string, boolean]) => {
        return <text onClick={(e) => {
            // TODO: stopProp or prevDef?
            // TODO: why val[0]?
            onClick && onClick(val[0])
        }}>{val[0]}</text>
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
    const [users, setUsers] = useState<Map<string, boolean>>(initial)
    useEffect(() => {
        setUsers(new Map(initial))
    }, [initial]);
    const update = (updated: Map<string, boolean>) => {
        setUsers(updated)
        updateParent && updateParent(updated)
    }
    const addNewUser = (uName: string) => {
        const defaultPerm = blanket || false
        if (users.size === 0) {/// TODO: ensure ok
            update(new Map<string, boolean>().set(uName, defaultPerm))
        } else {
            update(new Map<string, boolean>(users).set(uName, defaultPerm)) // TODO: validate ok here and in projects
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
                    <text>{"Can " + (canWrite ? "edit" : "view")}</text>
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
    // TODO: UPDATE TO STORE ACL LOCALLY, THEN UPDATE PARENT
    const [val, setVal] = useState(inp.ACL)
    useEffect(() => {
        setVal(inp.ACL)
    }, [inp.ACL])
    const updateBlanket = (b?: boolean) => {
        if (val !== undefined) {
            const temp = structuredClone(val)
            temp.blanketPerm = b
            setVal(temp)
        } else {
            setVal({ // TODO: ensure works properly
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
        if (a.blanketPerm === true) {
            return "Publicly Editable"
        } else if (a.blanketPerm === undefined) {
            return "Private"
        }
        return "Publicly Viewable"
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
                        updateParent={(s) => {
                            inp.updateParent(strToPerm(s))
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

export function CloneAcl(acl: ACL): ACL { // TODO: use?
    return {
        blanketPerm: acl.blanketPerm,
        users: (acl.users !== undefined && acl.users.size !== 0) ? new Map<string, boolean>(acl.users) : new Map<string, boolean>(),
        projects: (acl.projects !== undefined && acl.projects.size !== 0) ? new Map<string, boolean>(acl.projects) : new Map<string, boolean>(),
    }
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
}) {
    const [current, setCurrent] = React.useState(inp.initial)
    useEffect(()=>{
        const updated = {
            blanketPerm: inp.initial.blanketPerm,
            users: mapFor(inp.initial.users),
            projects: mapFor(inp.initial.projects),
        }
        setCurrent(updated)
    },[inp.initial])
    const cloneCurrent = ()=>{
        return structuredClone(current)
    }
    const update = (updated: ACL)=>{
        inp.updateParent(updated) // TODO: ensure ok
        setCurrent(updated)

    }
    const updateProjects = (updated: Map<string,boolean>)=>{
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
    const mapFor = (inpMap?: Map<string, boolean>):Map<string, boolean> => {
        return ((inpMap !== undefined && inpMap.size !== 0) ? (new Map<string, boolean>(inpMap)) : new Map<string, boolean>())
    }
    const depth = useContext(DepthContext)
    if (inp.readonly) {
        return <div>{/* TODO: TURN INTO A TABLE!!!!*/}
            <TestAndValidate todos={["preloaded values arent sticking around when updating.... troubleshoot"]}>
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
        <AclProjectsDisplay readonly={false} initial={inp.initial.projects || new Map()} onUsernClick={() => {
            // TODO: this?
        }} updateParent={updateProjects}/>
    </div>
}

export function DefaultAclDisplay({readonly, initial, updateParent}: {
    readonly: boolean,
    initial: ACL,
    updateParent: (acl: ACL) => void
}) {
    return <div>
        {"Default Access Control List"}
        <AclDisplay readonly={readonly} initial={initial} updateParent={updateParent}/>
    </div>
}